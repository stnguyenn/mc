/*
 * MinIO Client (C) 2018 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/minio/cli"
	json "github.com/minio/mc/pkg/colorjson"
	"github.com/minio/mc/pkg/console"
	"github.com/minio/mc/pkg/probe"
)

var adminUserAddCmd = cli.Command{
	Name:   "add",
	Usage:  "add a new user",
	Action: mainAdminUserAdd,
	Before: setGlobalsFromContext,
	Flags:  globalFlags,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} TARGET ACCESSKEY SECRETKEY

ACCESSKEY:
  Also called as username.

SECRETKEY:
  Also called as password.

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Add a new user 'foobar' to MinIO server.
     {{.DisableHistory}}
     {{.Prompt}} {{.HelpName}} myminio foobar foo12345
     {{.EnableHistory}}
`,
}

// checkAdminUserAddSyntax - validate all the passed arguments
func checkAdminUserAddSyntax(ctx *cli.Context) {
	if len(ctx.Args()) != 3 {
		cli.ShowCommandHelpAndExit(ctx, "add", 1) // last argument is exit code
	}
}

// userMessage container for content message structure
type userMessage struct {
	op         string
	Status     string   `json:"status"` // TODO: remove this?
	AccessKey  string   `json:"accessKey,omitempty"`
	SecretKey  string   `json:"secretKey,omitempty"`
	PolicyName string   `json:"policyName,omitempty"`
	UserStatus string   `json:"userStatus,omitempty"`
	MemberOf   []string `json:"memberOf,omitempty"`
}

func (u userMessage) String() string {
	switch u.op {
	case "list":
		userFieldMaxLen := 9
		accessFieldMaxLen := 20
		policyFieldMaxLen := 20

		// Create a new pretty table with cols configuration
		return newPrettyTable("  ",
			Field{"UserStatus", userFieldMaxLen},
			Field{"AccessKey", accessFieldMaxLen},
			Field{"PolicyName", policyFieldMaxLen},
		).buildRow(u.UserStatus, u.AccessKey, u.PolicyName)
	case "info":
		return console.Colorize("UserMessage", strings.Join(
			[]string{
				fmt.Sprintf("AccessKey: %s", u.AccessKey),
				fmt.Sprintf("Status: %s", u.UserStatus),
				fmt.Sprintf("PolicyName: %s", u.PolicyName),
				fmt.Sprintf("MemberOf: %s", strings.Join(u.MemberOf, ",")),
			}, "\n"))
	case "remove":
		return console.Colorize("UserMessage", "Removed user `"+u.AccessKey+"` successfully.")
	case "disable":
		return console.Colorize("UserMessage", "Disabled user `"+u.AccessKey+"` successfully.")
	case "enable":
		return console.Colorize("UserMessage", "Enabled user `"+u.AccessKey+"` successfully.")
	case "add":
		return console.Colorize("UserMessage", "Added user `"+u.AccessKey+"` successfully.")
	}
	return ""
}

func (u userMessage) JSON() string {
	u.Status = "success"
	jsonMessageBytes, e := json.MarshalIndent(u, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(jsonMessageBytes)
}

// mainAdminUserAdd is the handle for "mc admin user add" command.
func mainAdminUserAdd(ctx *cli.Context) error {
	checkAdminUserAddSyntax(ctx)

	console.SetColor("UserMessage", color.New(color.FgGreen))

	// Get the alias parameter from cli
	args := ctx.Args()
	aliasedURL := args.Get(0)

	// Create a new MinIO Admin Client
	client, err := newAdminClient(aliasedURL)
	fatalIf(err, "Unable to initialize admin connection.")

	fatalIf(probe.NewError(client.AddUser(args.Get(1), args.Get(2))).Trace(args...), "Cannot add new user")

	printMsg(userMessage{
		op:         "add",
		AccessKey:  args.Get(1),
		SecretKey:  args.Get(2),
		UserStatus: "enabled",
	})

	return nil
}
