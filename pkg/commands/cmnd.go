/*
Copyright 2016 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/
package commands

import (
	"github.com/spf13/cobra"
	"github.com/v3io/v3cli/sdk"
)

func NewCmdGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get [container-name] [path]",
		Short:   "Retrive object content",
		Long:    GetLongHelp(""),
		Example: GetExample(""),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sdk.RunGet(cmd.OutOrStdout(), Url, Container, Path, Verbose)
		},
	}

	AddWatch(cmd)
	return cmd
}

func NewCmdPut() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "put [container-name] [path]",
		Short:   "Upload object content from input file or stdin",
		Long:    GetLongHelp(""),
		Example: GetExample(""),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sdk.RunPut(cmd.OutOrStdout(), Url, Container, Path, InFile, Verbose)
		},
	}
	AddInFile(cmd)
	return cmd
}

func NewCmdPutitem() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "putitem [container-name] [table-path/key]",
		Short:   "Upload record content/fields from json input file or stdin",
		Long:    GetLongHelp(""),
		Example: GetExample(""),
		Aliases: []string{"puti"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return sdk.RunPutItem(cmd.OutOrStdout(), Url, Container, Path, InFile, Verbose)
		},
	}
	return cmd
}

func NewCmdGetitem() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "getitem [container-name] [table-path/key]",
		Short:   "Retrive record content/fields (as json struct)",
		Long:    GetLongHelp(""),
		Example: GetExample(""),
		Aliases: []string{"gi"},
		RunE: func(cmd *cobra.Command, args []string) error {
			attrs, _ := cmd.Flags().GetString("attrs")
			return sdk.RunGetItem(cmd.OutOrStdout(), Url, Container, Path, attrs, Verbose)
		},
	}
	cmd.Flags().StringP("attrs", "a", "*", "GetItem(s) Columns to return seperated by ','")
	return cmd
}

func NewCmdGetrecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "getrecords [container-name] [stream-path/shard-id] [-s seek][-t time][-n seq][-m max][-w int]",
		Short:   "Retrive one or more stream records",
		Long:    GetLongHelp(""),
		Example: GetExample(""),
		Aliases: []string{"gr"},
		RunE: func(cmd *cobra.Command, args []string) error {
			seek, _ := cmd.Flags().GetString("seek")
			maxrec, _ := cmd.Flags().GetInt("max-rec")
			interval, _ := cmd.Flags().GetInt("watch")
			return sdk.RunGetRecords(cmd.OutOrStdout(), Url, Container, Path, seek, maxrec, interval, Verbose)
		},
	}

	cmd.Flags().StringP("seek", "s", "EARLIEST", "Relative stream location [EARLIEST | LATEST | SEQUENCE | TIME]")
	cmd.Flags().StringP("time", "t", "", "Starting time - for TIME seek")
	cmd.Flags().IntP("sequence", "n", 0, "Starting sequence - for SEQUENCE seek")
	cmd.Flags().IntP("max-rec", "m", 50, "Max Records/Items to get per call")
	AddWatch(cmd)
	return cmd
}

func NewCmdPutrecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "putrecord [container-name] [stream-path]",
		Short:   "Upload stream record/message content from input file or stdin",
		Long:    GetLongHelp(""),
		Example: GetExample(""),
		Aliases: []string{"putr"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return sdk.RunPutRecord(cmd.OutOrStdout(), Url, Container, Path, InFile, Verbose)
		},
	}
	return cmd
}

func NewCmdCreatestream() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "createstream [container-name] [stream-path]",
		Short:   "Create a new stream with N shards",
		Long:    GetLongHelp(""),
		Example: GetExample(""),
		Aliases: []string{"cstr"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shards, _ := cmd.Flags().GetInt("shards")
			size, _ := cmd.Flags().GetInt("shardsize")
			return sdk.RunCreateStream(cmd.OutOrStdout(), Url, Container, Path, shards, size, Verbose)
		},
	}

	cmd.Flags().IntP("shards", "n", 1, "Number of shards (partitions)")
	cmd.Flags().IntP("shardsize", "s", 10, "Stream shard size in MB")
	cmd.Flags().IntP("retention", "r", 7, "Stream retention time in days")
	return cmd
}