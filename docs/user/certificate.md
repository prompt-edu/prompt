---
sidebar_position: 7
---

# 📜 Certificate Course Phase

The **Certificate Phase** allows instructors to generate and distribute course completion certificates to students. Certificates are generated as PDF documents from customizable [Typst](https://typst.app/) templates.

---

## Overview

The certificate phase provides:

- **Template Management**: Upload and manage Typst-based certificate templates
- **Preview**: Test certificates with mock data before releasing to students
- **Controlled Release**: Set a release date or release certificates immediately
- **Download Tracking**: Monitor which students have downloaded their certificates

---

## For Instructors

### Setting Up a Certificate Template

1. **Navigate** to the certificate phase in your course management view.
2. **Upload a Template**: On the settings page, paste or write your Typst template in the editor and save.
3. **Test the Template**: Click **Test Certificate** to generate a preview with mock data. This verifies your template compiles without errors.
4. **Set a Release Date** (optional): Use the release date picker to schedule when students can download certificates. Alternatively, click **Release Now** to make them available immediately.

### Template Format

Certificate templates use the [Typst](https://typst.app/) markup language. Your template receives student data via a JSON file.

#### Loading Data

```typst
#let data = json("data.json")
```

#### Available Fields

| Field | Type | Description |
|-------|------|-------------|
| `studentName` | string | Student's full name (first + last) |
| `courseName` | string | Course name from PROMPT |
| `teamName` | string | Team name (empty string if not allocated) |
| `date` | string | Generation date (e.g., "January 2, 2006") |

#### Example Template

```typst
#let data = json("data.json")

#set page(paper: "a4")
#set text(font: "Linux Libertine")

#align(center + horizon)[
  #text(size: 28pt, weight: "bold")[Certificate of Completion]

  #v(2em)

  #text(size: 18pt)[This certifies that]

  #v(1em)

  #text(size: 24pt, weight: "bold")[#data.studentName]

  #v(1em)

  #text(size: 18pt)[has successfully completed the course]

  #v(1em)

  #text(size: 22pt, style: "italic")[#data.courseName]

  #if data.teamName != "" [
    #v(1em)
    #text(size: 16pt)[as a member of team #data.teamName]
  ]

  #v(3em)

  #text(size: 14pt)[#data.date]
]
```

#### Embedding Images

Since templates are stored as a single file, images (logos, signatures, etc.) must be embedded directly in the template. No external files can be referenced.

**SVG graphics** (best for logos, icons, and decorative elements — they scale without losing quality):

```typst
#let logo = bytes("<svg xmlns='http://www.w3.org/2000/svg' width='200' height='100'>...</svg>")
#image(logo, width: 4cm)
```

**Raster images (PNG/JPG)** — encode the image as base64 and wrap it in an SVG:

```typst
#let signature = bytes("<svg xmlns='http://www.w3.org/2000/svg' width='400' height='100'><image href='data:image/png;base64,iVBORw0KGgo...' width='400' height='100'/></svg>")
#image(signature, width: 5cm)
```

To convert a PNG file to a base64 string, run in your terminal:

```bash
base64 -i image.png | tr -d '\n'
```

:::tip
Set the SVG `width` and `height` attributes to the original pixel dimensions of your image for a correct aspect ratio. Use the `width` parameter on `#image()` to control the display size in the certificate.
:::

#### Template Tips

- Target **A4 page format** for consistency
- Use `json("data.json")` to access certificate data
- Test with the **Test Certificate** button before releasing — it shows compilation errors if the template is invalid
- Fonts must be system fonts available in the deployment environment
- Images must be embedded inline (see above) — external file references will not work
- See the [Typst documentation](https://typst.app/docs) for the full language reference

### Managing Participants

The **Participants** tab shows all enrolled students with:

- **Download Status**: Whether a student has downloaded their certificate
- **Download Count**: How many times a student has downloaded
- **First/Last Download**: Timestamps for download tracking

Instructors can also download a certificate for any individual student from the participants table.

### Release Configuration

You have three options for releasing certificates:

- **Set a Release Date**: Choose a future date/time when certificates become available
- **Release Now**: Make certificates immediately available to all students
- **Clear**: Remove the release date (certificates become unavailable until a new date is set)

:::tip
Use the **Test Certificate** button before setting a release date to ensure your template works correctly.
:::

---

## For Students

### Downloading Your Certificate

1. **Navigate** to the certificate phase in your course dashboard.
2. If certificates have been released, you will see a **Download Certificate** button.
3. Click the button to download your personalized PDF certificate.
4. You can download your certificate multiple times — the download counter tracks this for the instructor.

### Certificate Availability

- Certificates are available only after the instructor has released them (either by setting a release date that has passed, or by releasing immediately).
- If certificates are not yet available, you will see a message indicating when they will be released.

:::note
If you encounter any issues downloading your certificate, contact your course instructor.
:::
