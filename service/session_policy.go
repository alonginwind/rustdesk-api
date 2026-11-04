package service

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lejianwen/rustdesk-api/v2/global"
	"gopkg.in/yaml.v3"
)

var sessionConfigMu sync.Mutex
var runtimeSessionLifetime atomic.Int64

type SessionPolicySettings struct {
	Duration           string `json:"duration"`
	EffectiveDuration  string `json:"effective_duration"`
	Revision           string `json:"revision"`
	Writable           bool   `json:"writable"`
	EnvironmentManaged bool   `json:"environment_managed"`
}

func sessionConfigFile() (string, error) {
	if global.Viper == nil || global.Viper.ConfigFileUsed() == "" {
		return "", errors.New("configuration file is unavailable")
	}
	return filepath.EvalSymlinks(global.Viper.ConfigFileUsed())
}

func sessionDurationNode(doc *yaml.Node, create bool) (*yaml.Node, error) {
	if len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("expected configuration mapping")
	}
	root := doc.Content[0]
	var app *yaml.Node
	for i := 0; i < len(root.Content); i += 2 {
		if root.Content[i].Value == "app" {
			app = root.Content[i+1]
			break
		}
	}
	if app == nil && create {
		app = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "app", Tag: "!!str"}, app)
	}
	if app == nil {
		return nil, nil
	}
	if app.Kind != yaml.MappingNode {
		return nil, errors.New("app must be a mapping")
	}
	for i := 0; i < len(app.Content); i += 2 {
		if app.Content[i].Value == "token-expire" {
			value := app.Content[i+1]
			if value.Kind != yaml.ScalarNode {
				return nil, errors.New("token-expire must be a scalar")
			}
			return value, nil
		}
	}
	if !create {
		return nil, nil
	}
	value := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str"}
	app.Content = append(app.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "token-expire"}, value)
	return value, nil
}

func (us *UserService) SessionPolicy() (*SessionPolicySettings, error) {
	sessionConfigMu.Lock()
	defer sessionConfigMu.Unlock()
	return us.readSessionPolicy()
}

func (us *UserService) readSessionPolicy() (*SessionPolicySettings, error) {
	path, err := sessionConfigFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	node, err := sessionDurationNode(&doc, false)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, errors.New("configuration must be a regular file")
	}
	value := ""
	if node != nil {
		value = node.Value
	}
	_, managed := os.LookupEnv("RUSTDESK_API_APP_TOKEN_EXPIRE")
	writable := false
	if !managed {
		if f, e := os.OpenFile(path, os.O_WRONLY, 0); e == nil {
			writable = true
			f.Close()
		}
	}
	return &SessionPolicySettings{Duration: value, EffectiveDuration: us.UserTokenLifetime().String(), Revision: fmt.Sprintf("%d:%d", st.ModTime().UnixNano(), st.Size()), Writable: writable, EnvironmentManaged: managed}, nil
}

// SaveSessionPolicy changes only app.token-expire in the configured YAML file.
// The server chooses the path; the request cannot supply a path or other fields.
func (us *UserService) SaveSessionPolicy(value, revision string) error {
	duration, err := time.ParseDuration(value)
	if err != nil || duration < time.Minute || duration > 3650*24*time.Hour {
		return errors.New("duration must be between one minute and 3650 days")
	}
	sessionConfigMu.Lock()
	defer sessionConfigMu.Unlock()
	current, err := us.readSessionPolicy()
	if err != nil {
		return err
	}
	if current.EnvironmentManaged {
		return errors.New("session lifetime is managed by the deployment environment")
	}
	if !current.Writable {
		return errors.New("configuration file is read-only")
	}
	if revision == "" || revision != current.Revision {
		return errors.New("configuration changed; reload before saving")
	}
	path, err := sessionConfigFile()
	if err != nil {
		return err
	}
	old, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(old, &doc); err != nil {
		return err
	}
	node, err := sessionDurationNode(&doc, true)
	if err != nil {
		return err
	}
	node.Value = value
	node.Tag = "!!str"
	var encoded bytes.Buffer
	encoder := yaml.NewEncoder(&encoded)
	encoder.SetIndent(2)
	if err = encoder.Encode(&doc); err != nil {
		return err
	}
	encoder.Close()
	var beforeValues, afterValues map[string]interface{}
	if err = yaml.Unmarshal(old, &beforeValues); err != nil {
		return err
	}
	if err = yaml.Unmarshal(encoded.Bytes(), &afterValues); err != nil {
		return err
	}
	// Anchors may make one scalar feed several settings. Refuse a change if
	// anything except the selected app.token-expire value would be affected.
	for _, values := range []map[string]interface{}{beforeValues, afterValues} {
		if app, ok := values["app"].(map[string]interface{}); ok {
			delete(app, "token-expire")
			if len(app) == 0 {
				delete(values, "app")
			}
		}
	}
	if !reflect.DeepEqual(beforeValues, afterValues) {
		return errors.New("editing token-expire would change other YAML values")
	}
	// Prepare the replacement before touching the live file. A read-only bind mount
	// or unwritable directory produces an error and leaves the original intact.
	f, err := os.CreateTemp(filepath.Dir(path), ".session-policy-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(st.Mode().Perm()); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(encoded.Bytes()); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	fresh, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	freshStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(fresh, old) || fmt.Sprintf("%d:%d", freshStat.ModTime().UnixNano(), freshStat.Size()) != revision {
		return errors.New("configuration changed; reload before saving")
	}
	backup := path + ".session-policy-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".bak"
	b, err := os.OpenFile(backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err = b.Write(old); err != nil {
		b.Close()
		return err
	}
	if err = b.Sync(); err != nil {
		b.Close()
		return err
	}
	if err = b.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	runtimeSessionLifetime.Store(int64(duration))
	return nil
}
