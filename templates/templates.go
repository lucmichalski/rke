package templates

import (
	"bytes"
	"encoding/json"
	"github.com/blang/semver"
	"github.com/rancher/kontainer-driver-metadata/rke/templates"
	"github.com/sirupsen/logrus"
	"text/template"

	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rke/metadata"
)

func CompileTemplateFromMap(tmplt string, configMap interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Funcs(template.FuncMap{"GetKubednsStubDomains": GetKubednsStubDomains}).Parse(tmplt))
	if err := t.Execute(out, configMap); err != nil {
		return "", err
	}
	return out.String(), nil
}

func GetVersionedTemplates(templateName string, data map[string]interface{}, k8sVersion string) string {
	if template, ok := data[templateName]; ok {
		return convert.ToString(template)
	}
	return getTemplate(templateName, k8sVersion)
}

func GetKubednsStubDomains(stubDomains map[string][]string) string {
	json, _ := json.Marshal(stubDomains)
	return string(json)
}

func GetDefaultVersionedTemplate(templateName string, data map[string]interface{}) string {
	if template, ok := data[templateName]; ok {
		return convert.ToString(template)
	}
	versionData := metadata.K8sVersionToTemplates[templateName]
	return metadata.K8sVersionToTemplates[templates.TemplateKeys][versionData["default"]]
}

func getTemplate(templateName, k8sVersion string) string {
	versionData := metadata.K8sVersionToTemplates[templateName]
	toMatch, err := semver.Make(k8sVersion[1:])
	if err != nil {
		logrus.Errorf("k8sVersion not sem-ver %s %v", k8sVersion, err)
		return metadata.K8sVersionToTemplates[templates.TemplateKeys][versionData["default"]]
	}
	for k := range versionData {
		if k == "default" {
			continue
		}
		testRange, err := semver.ParseRange(k)
		if err != nil {
			logrus.Errorf("range for %s not sem-ver %v %v", templateName, testRange, err)
			continue
		}
		if testRange(toMatch) {
			return metadata.K8sVersionToTemplates[templates.TemplateKeys][versionData[k]]
		}
	}
	return metadata.K8sVersionToTemplates[templates.TemplateKeys][versionData["default"]]
}
