---
title: Orgs
sidebar_position: 52
---

# Orgs

An org is an organizational unit such as a university chair. Orgs can be nested, so the same model
fits a single chair, several chairs on one instance, and a whole university. Courses without an org
keep working as before.

Open **Admin → Orgs** in the management console. The pages are only available to PROMPT
administrators.

| Field             | Meaning                                                                                                   |
| ----------------- | --------------------------------------------------------------------------------------------------------- |
| **Name**          | The display name. It can be renamed at any time and does not have to be unique.                           |
| **Slug**          | The org's identifier, such as `aet`. It names the org's Keycloak roles and cannot be changed later.       |
| **Parent Org**    | Optional. Places the org below another org.                                                               |
| **School**        | Optional, such as the school or faculty the org belongs to.                                               |
| **University**    | Optional.                                                                                                 |
| **Website**       | Optional. Must be an absolute `http` or `https` URL.                                                      |
| **Contact Email** | Optional. A plain address such as `chair@example.com`, without a display name.                            |

## Slug rules

The slug is filled in from the name while you type and can be edited before the org is created. It
must be 2–50 characters of lowercase letters and digits, separated by single hyphens, such as
`applied-education-technologies`. It must be unique across all orgs, including nested ones, and
must not contain a `cg` segment, which is reserved for the custom group roles of courses.

## Keycloak roles

Creating an org also creates the Keycloak group `/Orgs/<slug>` with the subgroups `Admin` and
`Member`, and the client roles `org-<slug>-Admin` and `org-<slug>-Member`. Nobody holds them at
first, not even the administrator who created the org.

## Nesting

Change an org's parent in the **Hierarchy** card of its page, or choose **None (top level)**. An
org cannot be placed below itself or one of its own child orgs. Nesting only records the structure
for now: the Admins and Members of a parent org get no rights on its child orgs.

## Editing and deleting

- **Edit** the name and the optional fields in the **Details** card. Leaving an optional field empty
  clears it.
- **Delete** an org in the **Danger Zone**. This also removes its Keycloak groups and roles and every
  membership. An org that still has child orgs or courses cannot be deleted: move or delete those
  first.
