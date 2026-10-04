---
title: Study Programs
sidebar_position: 50
---

# Study Programs

Applicants pick their study program from a list that PROMPT administrators manage centrally. The
same list is used for every course. Applicants whose program is not on the list choose **Other** and
type it in.

Open **Admin → Study Programs** in the management console. The page is only available to PROMPT
administrators.

| Column         | Meaning                                                                                                                    |
| -------------- | -------------------------------------------------------------------------------------------------------------------------- |
| **Name**       | What applicants see in the application form. Must be unique, ignoring case and surrounding spaces.                         |
| **Short Name** | Optional label for charts, such as `CS` for Computer Science. The name is used otherwise. Each chart label must be unique. |
| **Students**   | How many students currently have exactly this program stored.                                                              |

## Adding, renaming, and removing

- **Add** a program with **Add Study Program**. It appears in the application form right away.
- **Rename** a program by editing it. Every student who has the old name is updated to the new one,
  so the rename does not move them into **Other**. The dialog shows how many students are affected.
- **Remove** a program from the list with the row's delete action. Students keep the program they
  entered; it then counts as **Other** in the application statistics, like any free-text program.

The names **Other** (free-text programs) and **Unknown** (applications without a study program) are
reserved for the application statistics and cannot be used as a name or short name.
