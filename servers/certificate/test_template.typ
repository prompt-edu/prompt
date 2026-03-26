// Self-contained certificate template for PROMPT 2.0
// The generator writes data.json in the same directory as this template.
// JSON fields: studentName, courseName, teamName, date

#let TUM_blue = rgb("#0065BD")
#let TUM_dark = rgb("#003359")
#let TUM_light = rgb("#98C6EA")

#let data = json("data.json")

#set page("a4", margin: (left: 25mm, right: 25mm, top: 30mm, bottom: 30mm), background: [
  // Top decorative bar
  #place(top + left, rect(width: 100%, height: 8mm, fill: TUM_blue))
  // Bottom decorative bar
  #place(bottom + left, rect(width: 100%, height: 8mm, fill: TUM_blue))
  // Subtle corner accent
  #place(top + right, dx: -20mm, dy: 8mm, rect(width: 60mm, height: 2mm, fill: TUM_light))
])

#set text(font: "New Computer Modern", size: 12pt)
#set par(first-line-indent: 0em)

// Title
#v(3cm)
#align(center)[
  #text(weight: "bold", size: 14pt, fill: TUM_blue, tracking: 4pt, upper("Technical University of Munich"))
]

#v(0.5cm)
#align(center)[
  #line(length: 60%, stroke: 0.5pt + TUM_light)
]

#v(1cm)
#align(center)[
  #text(weight: "regular", size: 36pt, fill: TUM_blue, upper("Certificate"))
]

#v(0.3cm)
#align(center)[
  #text(size: 14pt, fill: TUM_dark, "of Completion")
]

// Body
#v(1.5cm)
#align(center)[
  #text(size: 13pt)[
    This is to certify that
  ]
]

#v(0.8cm)
#align(center)[
  #text(weight: "bold", size: 26pt, fill: TUM_blue, upper(data.studentName))
]

#v(0.8cm)
#align(center)[
  #text(size: 13pt)[
    has successfully completed the course
  ]
]

#v(0.5cm)
#align(center)[
  #text(weight: "bold", size: 22pt, fill: TUM_blue, upper(data.courseName))
]

// Show team name if available
#if data.teamName != "" [
  #v(0.3cm)
  #align(center)[
    #text(size: 13pt)[
      in collaboration with
    ]
  ]
  #v(0.2cm)
  #align(center)[
    #text(weight: "bold", size: 18pt, fill: TUM_blue, upper(data.teamName))
  ]
]

#v(0.5cm)
#align(center)[
  #text(size: 13pt)[
    demonstrating dedication, skill, and commitment to excellence
    in software engineering practices.
  ]
]

// Date and signature
#v(2fr)
#align(center)[
  #text(size: 12pt)[
    Munich, #data.date
  ]
]

#v(1.5cm)
#align(center)[
  #line(length: 40%, stroke: 0.5pt + TUM_dark)
  #v(-0.2cm)
  #text(size: 11pt, fill: TUM_dark)[
    Prof. Dr. Stephan Krusche
  ]
]
