package kube2cdk8s

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/pulumi/kube2pulumi/pkg/kube2pulumi"
)

func Kube2CDK8S(filePath string) (string, error) {

	path, _, err := kube2pulumi.Kube2PulumiFile(filePath, "typescript")
	if err != nil {
		return "", err
	}

	input, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "import") {
			lines[i] = ``
		}
	}
	output := strings.Join(lines, "\n")

	re := regexp.MustCompile("(?m)[\r\n]+^.*const.*$")
	res := re.ReplaceAllString(output, `new cdk8s.ApiObject("", this, {`)

	defer os.Remove(path)

	return res, nil
}
