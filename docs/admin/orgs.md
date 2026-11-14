---
title: Orgs
sidebar_position: 52
---

# Orgs

An org is an organizational unit such as a university chair. Orgs can be nested, so the same model
fits a single chair, several chairs on one instance, and a whole university. Courses without an org
keep working as before.

PROMPT administrators manage every org under **Admin → Orgs** in the management console. The Admins
and Members of an org open its page from **My Org** in the sidebar instead, see [Members](#members).

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
first, not even the administrator who created the org: add the first Admin in the **Members** card.

## Nesting

Change an org's parent in the **Hierarchy** card of its page, or choose **None (top level)**. An
org cannot be placed below itself or one of its own child orgs. Nesting only records the structure
for now: the Admins and Members of a parent org get no rights on its child orgs.

## Members

An org has Admins and Members: the users in its Keycloak groups `/Orgs/<slug>/Admin` and
`/Orgs/<slug>/Member`. PROMPT keeps no copy of them, so the **Members** card reads and changes these
groups directly. PROMPT administrators and the org's Admins see the card; Members see only the org's
details, read-only.

- **Add** a user by their university login and choose their role. A user without a university login
  in Keycloak is found by their username. Adding a member again changes their role, since a user
  holds one role in an org at a time.
- **Change** a member's role or **remove** them from the org in their row.
- Nobody can change their own membership, so an Admin can neither remove themselves nor give up the
  Admin role. Ask another Admin of the org or a PROMPT administrator.
- A change takes effect once the user's session renews, at the latest at their next login.

## Editing and deleting

- **Edit** the name and the optional fields in the **Details** card. PROMPT administrators and the
  org's Admins can do this. Leaving an optional field empty clears it.
- **Delete** an org in the **Danger Zone**. This also removes its Keycloak groups and roles and every
  membership. An org that still has child orgs or courses cannot be deleted: move or delete those
  first.

## When deleting fails

A failed delete shows an error and keeps the org. In most cases its Keycloak groups and members are
unchanged, so you can simply delete it again.

If Keycloak had already removed the org's groups when the delete failed, PROMPT recreates the groups
and roles and puts the org's Admins and Members back. Only when a member cannot be re-added does the
error say so, and the server log names the users who are missing. Either delete the org again, or
keep it and add them again in the **Members** card or to the `/Orgs/<slug>/Admin` and
`/Orgs/<slug>/Member` groups in Keycloak.
