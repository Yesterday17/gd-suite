package touch

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/flags"
	"github.com/rclone/rclone/fs/fspath"
	"github.com/rclone/rclone/fs/object"
	"github.com/spf13/cobra"
)

var (
	notCreateNewFile bool
	timeAsArgument   string
	localTime        bool
	FirstOnly        bool
)

const (
	defaultLayout          string = "060102"
	layoutDateWithTime            = "2006-01-02T15:04:05"
	layoutDateWithTimeNano        = "2006-01-02T15:04:05.999999999"
)

func init() {
	cmd.Root.AddCommand(commandDefinition)
	cmdFlags := commandDefinition.Flags()
	flags.BoolVarP(cmdFlags, &notCreateNewFile, "no-create", "C", false, "Do not create the file if it does not exist.")
	flags.StringVarP(cmdFlags, &timeAsArgument, "timestamp", "t", "", "Use specified time instead of the current time of day.")
	flags.BoolVarP(cmdFlags, &localTime, "localtime", "", false, "Use localtime for timestamp, not UTC.")
	flags.BoolVarP(cmdFlags, &FirstOnly, "first-only", "F", false, "Do not touch parent folders.")
}

type touchJob struct {
	fs       fs.Fs
	fileName string
}

var commandDefinition = &cobra.Command{
	Use:   "touch remote:path",
	Short: `Create new file or change file modification time.`,
	Long: `
Set the modification time on object(s) as specified by remote:path to
have the current time.

If remote:path does not exist then a zero sized object will be created
unless the --no-create flag is provided.

If --timestamp is used then it will set the modification time to that
time instead of the current time. Times may be specified as one of:

- 'YYMMDD' - e.g. 17.10.30
- 'YYYY-MM-DDTHH:MM:SS' - e.g. 2006-01-02T15:04:05
- 'YYYY-MM-DDTHH:MM:SS.SSS' - e.g. 2006-01-02T15:04:05.123456789

Note that --timestamp is in UTC if you want local time then add the
--localtime flag.
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)

		var path = args[0]
		var err error
		var toTouch []touchJob
		for {
			if strings.HasSuffix(path, "/") {
				// remove padding /
				path = path[:len(path)-1]
			}

			dstRemote, file, err := fspath.Split(path)
			if err != nil || dstRemote == path {
				break
			}
			if dstRemote == "" {
				dstRemote = "."
			}
			fsdt := cmd.NewFsDir([]string{dstRemote})
			toTouch = append(toTouch, touchJob{fsdt, file})

			if FirstOnly {
				break
			}
			path = dstRemote
		}

		cmd.Run(true, false, command, func() error {
			for _, t := range toTouch {
				err = Touch(context.Background(), t.fs, t.fileName)
				if err != nil {
					return err
				}
			}
			return nil
		})
	},
}

// Touch create new file or change file modification time.
func Touch(ctx context.Context, fsrc fs.Fs, srcFileName string) (err error) {
	timeAtr := time.Now()
	if timeAsArgument != "" {
		layout := defaultLayout
		if len(timeAsArgument) == len(layoutDateWithTime) {
			layout = layoutDateWithTime
		} else if len(timeAsArgument) > len(layoutDateWithTime) {
			layout = layoutDateWithTimeNano
		}
		var timeAtrFromFlags time.Time
		if localTime {
			timeAtrFromFlags, err = time.ParseInLocation(layout, timeAsArgument, time.Local)
		} else {
			timeAtrFromFlags, err = time.Parse(layout, timeAsArgument)
		}
		if err != nil {
			return errors.Wrap(err, "failed to parse date/time argument")
		}
		timeAtr = timeAtrFromFlags
	}
	file, err := fsrc.NewObject(ctx, srcFileName)
	if err != nil {
		if !notCreateNewFile {
			var buffer []byte
			src := object.NewStaticObjectInfo(srcFileName, timeAtr, int64(len(buffer)), true, nil, fsrc)
			_, err = fsrc.Put(ctx, bytes.NewBuffer(buffer), src)
			if err != nil {
				return err
			}
		}
		if err == fs.ErrorNotAFile && file != nil {
			err = file.SetModTime(ctx, timeAtr)
			if err != nil {
				return errors.Wrap(err, "touch: couldn't set mod time of folder")
			}
			return nil
		}
		return nil
	}
	err = file.SetModTime(ctx, timeAtr)
	if err != nil {
		return errors.Wrap(err, "touch: couldn't set mod time")
	}
	return nil
}
