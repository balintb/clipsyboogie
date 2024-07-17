package main

// Launchd template
func Template() string {
	return `
<?xml version="1.0" encoding="UTF-8"?>
 <!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd" >
 <plist version='1.0'>
   <dict>
     <key>Label</key><string>{{.Label}}</string>
	 <key>ProgramArguments</key>
		<array>
			<string>{{.Program}}</string>
			<string>listen</string>
			<string>--interval</string>
			<string>{{.Interval}}</string>
		</array>
     <key>StandardOutPath</key><string>/tmp/{{.Label}}.out.log</string>
     <key>StandardErrorPath</key><string>/tmp/{{.Label}}.err.log</string>
     <key>KeepAlive</key><{{.KeepAlive}}/>
     <key>RunAtLoad</key><{{.RunAtLoad}}/>
   </dict>
</plist>
`
}
