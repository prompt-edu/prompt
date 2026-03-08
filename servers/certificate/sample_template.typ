#let vars = json("vars.json")

#set page(margin: 1cm)
#set text(size: 12pt, font: "Times New Roman")

#align(center)[
  #text(size: 24pt, weight: "bold")[Certificate of Completion]

  #v(1cm)

  #text(size: 16pt)[This certifies that]

  #v(0.5cm)

  #text(size: 20pt, weight: "bold")[#vars.studentName]

  #v(0.5cm)

  #text(size: 16pt)[has successfully completed the course]

  #v(0.5cm)

  #text(size: 18pt, weight: "bold")[#vars.courseName]

  #v(0.5cm)

  #text(size: 14pt)[as a member of #vars.teamName]

  #v(1cm)

  #text(size: 12pt)[Date: #vars.date]
]
