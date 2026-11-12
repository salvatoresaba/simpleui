# Timeline

A vertical list of dated events with a status dot (done, active, pending or error) and optional rich content.

## Complete source

```sui
css ".timeline-page \\{ max-width: 720px; margin: 0 auto; padding: 28px; display: flex; flex-direction: column; gap: 14px; \\} .timeline-note \\{ color: #8fa0b8; font-size: 13px; \\}"

view main {
	title "Timeline"
	col class: "timeline-page" {
		h1 "Timeline"
		p "Events with a date, a status and optional rich content." class: "timeline-note"
		timeline {
			timelineitem "Project created" date: "10 Nov, 09:12" status: "done"
			timelineitem "Build completed" date: "10 Nov, 09:30" status: "done" {
				p "Artifacts are ready for review."
			}
			timelineitem "Review" date: "11 Nov, 14:02" status: "active" {
				p "Waiting for approval."
			}
			timelineitem "Deploy" date: "12 Nov, 08:00" status: "pending"
		}
	}
}
```

## Designer

Choose **Content > Timeline** in the palette. Each `timelineitem` exposes **date** and **status** in the **Timeline event** section; add children for extra content.

## Build

```bash
simpleui check app.sui --json
simpleui build app.sui --out dist
simpleui build app.sui --target desktop --out dist
```

