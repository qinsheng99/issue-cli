package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"sort"

	"github.com/spf13/cobra"

	"github.com/opensourceways/issue-cli/util"
)

const basefile = "%s.txt"

type issueTypeOption struct {
	Streams
	h *util.Request

	name string
	file bool
}

var issueExample = `
  # Get all issue type list
  issue-cli get issue-type or issue get it

  # Get issue template(default print os.Stdout)
  issue-cli get it -n [name]
`

func newIssueTypeOption(s Streams) *issueTypeOption {
	return &issueTypeOption{Streams: s, h: util.NewRequest(nil)}
}

func newIssueTypeCmd(s Streams) *cobra.Command {
	o := newIssueTypeOption(s)

	cmd := &cobra.Command{
		Use:     "issue_type",
		Aliases: []string{"it"},
		Short:   "get openeuler community issue type",
		Example: issueExample,
		Run: func(cmd *cobra.Command, args []string) {
			checkErr(o.Validate(cmd, args))
			checkErr(o.Run())
		},
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", o.name, "issue type name use to obtain a issue template")
	cmd.Flags().BoolVarP(&o.file, "file", "f", o.file, "If true, output the content to a file")

	return cmd
}

func (i *issueTypeOption) Validate(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return util.UsageErrorf(cmd, "unexpected args: %v", args)
	}
	return nil
}

func (i *issueTypeOption) Run() error {
	if len(i.name) > 0 {
		return i.uniqueOne()
	}

	var res = struct {
		baseResp
		Data []struct {
			UniqueId int64  `json:"id"`
			Name     string `json:"name"`
		}
	}{}

	_, err := i.h.CustomRequest(IssueTypeUrl, "GET", nil, nil, nil, &res)
	if err != nil {
		return err
	}

	if res.Code != 200 {
		return fmt.Errorf(res.Msg)
	}

	err = i.printContextHeaders(i.Out)
	if err != nil {
		return err
	}
	var data = res.Data
	sort.Slice(data, func(i, j int) bool {
		return data[i].UniqueId < data[j].UniqueId
	})
	for _, v := range data {
		_, err = fmt.Fprintf(i.Out, "%-15d\t%s\n", v.UniqueId, v.Name)
	}

	return err
}

func (i *issueTypeOption) uniqueOne() error {
	var v = url.Values{}
	v.Add("name", i.name)
	var res = struct {
		baseResp
		Data []struct {
			UniqueId int64  `json:"id"`
			Name     string `json:"name"`
			Template string `json:"template"`
		}
	}{}

	_, err := i.h.CustomRequest(IssueTypeUrl, "GET", nil, nil, v, &res)
	if err != nil {
		return err
	}
	if res.Code != 200 {
		return fmt.Errorf(res.Msg)
	}

	if len(res.Data) == 0 {
		return fmt.Errorf("not found issue type : %s", i.name)
	}

	if i.file {
		return i.writeFile(res.Data[0].Template)
	}

	_, err = fmt.Fprintln(i.Out, res.Data[0].Template)
	return err
}

func (i *issueTypeOption) writeFile(content string) error {
	var file = fmt.Sprintf(basefile, "issue")

	return ioutil.WriteFile(file, []byte(content), fs.ModePerm)
}

func (i *issueTypeOption) printContextHeaders(out io.Writer) error {
	columnNames := []any{"UNIQUEID", "NAME"}
	_, err := fmt.Fprintf(out, "%-15s\t%s\n", columnNames...)
	return err
}
