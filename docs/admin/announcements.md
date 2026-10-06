---
title: Announcements
sidebar_position: 55
---

# Announcements

Announcements are banners at the top of every page, including the public application pages. Use
them for maintenance windows, outages, or other notices that every user should see.

Open **Admin → Announcements** in the management console. The page is only available to PROMPT
administrators.

## Creating an announcement

**Create Announcement** opens a form with a live preview of the banner.

| Field          | Meaning                                                                               |
| -------------- | ------------------------------------------------------------------------------------- |
| **Severity**   | **Info**, **Warning**, or **Critical**. Sets the color and icon of the banner.        |
| **Title**      | Optional bold prefix, up to 100 characters.                                           |
| **Message**    | The text of the banner, up to 300 characters. Required.                               |
| **Link URL**   | Optional `http` or `https` link shown after the message. Opens in a new tab.          |
| **Link label** | Optional text for the link, up to 50 characters. Defaults to **Learn more**.          |
| **Starts**     | Optional time the banner appears. Leave empty to show it as soon as it is enabled.    |
| **Expires**    | Optional time the banner disappears. Must be after the start. Leave empty to keep it. |
| **Enabled**    | Whether the banner is shown at all.                                                   |

New announcements start **disabled**, so you can prepare one before it goes live. Enable it in the
form or with the **Enabled** switch in the table.

## Status

The table shows each announcement's status:

| Status        | Meaning                                                  |
| ------------- | -------------------------------------------------------- |
| **Active**    | Enabled, started, and not expired. Users see the banner. |
| **Scheduled** | Enabled, but the start time is still in the future.      |
| **Disabled**  | Turned off. Users do not see the banner.                 |
| **Expired**   | The expiry time has passed. Users do not see the banner. |

Expired announcements are hidden from the table unless **Show expired** is on. Open browsers pick up
changes within five minutes, or as soon as the tab regains focus.

## What users see

When several announcements are active, users see one banner at a time and can page through them.

Users can dismiss a banner. The dismissal is stored in their browser only, so the banner comes back
in another browser or after clearing site data. Editing an announcement shows it again to everyone
who dismissed it, so use an edit for updates that users must not miss.
