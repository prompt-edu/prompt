---
sidebar_position: 6
---

# Reusable UI Components

The shared library provides a range of reusable UI components you can easily incorporate into your code. It lives in the [`prompt-edu/prompt-lib`](https://github.com/prompt-edu/prompt-lib) repository and is consumed here as the `@tumaet/prompt-ui-components` package, so every component below is imported from that package.

---


## ManagementPageHeader

The `ManagementPageHeader` component is a simple but essential part of the user interface. It is used to render a prominent page title or header for management-related pages, making it clear what the purpose of the page is. This component provides a clean, consistent look for headers across different pages in the application.

### Key Features

- **Customizable Title**: The component allows any content to be passed as a child, enabling flexibility to display different titles or headings as needed.
- **Styling**: The header is styled with a large font size (`text-4xl`), bold font (`font-bold`), and a margin at the bottom (`mb-6`), ensuring it stands out on the page.

### Usage

- **Props**:
  - `children`: The content (typically a title or header text) to be displayed inside the `h1` element. This can be a string, a component, or any other valid React node.

The `ManagementPageHeader` component helps standardize page headers across your application and ensures they are easily readable and visually appealing.

(errorpage)=

## ErrorPage

The `ErrorPage` component is designed to handle and display error messages in a user-friendly manner. It provides an alert box with a title, description, and a customizable message to inform users of issues such as network failures or server errors. Additionally, it offers options to retry the action that caused the error or log out of the application.

### Key Features

- **Title & Description**: The component allows customization of the error title and description. By default, it displays a generic "Error" message with a description explaining the issue.
- **Message**: Displays a detailed error message that can be customized. The default message informs users of potential network issues and suggests retrying or waiting.
- **Retry Button**: If the `onRetry` function is provided, a retry button is displayed, allowing the user to attempt the action again.
- **Logout Button**: If the `onLogout` function is provided, a logout button is shown, allowing users to log out of the application if needed.

### Usage

- **Props**:
  - `title` (optional): The title of the error message (defaults to "Error").
  - `description` (optional): A short description of the error (defaults to "An error occurred").
  - `message` (optional): A detailed message explaining the error (defaults to a generic message).
  - `onRetry` (optional): A function that is called when the retry button is clicked.
  - `onLogout` (optional): A function that is called when the logout button is clicked.

This component ensures that errors are presented in a clear and accessible manner, with the flexibility to add custom actions (such as retrying or logging out) based on the specific needs of the application.

## CoursePhaseMailing

The `CoursePhaseMailing` component is used to manage the mailing functionality within a specific course phase.
To integrate the `CoursePhaseMailing` component and set up a Mailing page correctly, follow these steps:

### 1. Import Necessary Components

Start by importing the essential components required for the page, including:

1. **UI Components**:

   - [`ManagementPageHeader`](#managementpageheader): A header component for management pages.
   - [`ErrorPage`](#errorpage): A component to display error messages with retry options.

2. **Routing Utilities**:

   - `useParams` (from React Router): Retrieves the `phaseId` from the URL.

3. **Data Fetching Tools**:

   - `useQuery` (from `@tanstack/react-query`): Handles data fetching and caching.

4. **API Query Function**:
   - `getCoursePhase`: Fetches course phase details using the `phaseId`. This function is exported from `@tumaet/prompt-shared-state` and returns a `CoursePhaseWithMetaData` interface, defined in the same package.

### 2. Retrieve Course Phase Data

- Extract the `phaseId` using `useParams`.
- Use `useQuery` to fetch the relevant course phase data from the core.
- Ensure the `queryFn` correctly calls `getCoursePhase` with the `phaseId`.

### 3. Handle Loading and Errors

- Display a loading indicator (e.g. `Loader2`) while the data is being fetched.
- Show an error page (`ErrorPage`) with a retry option if the request fails.

### 4. Render the Mailing Page

- Use `ManagementPageHeader` for the page title.
- Pass the retrieved `coursePhase` data into the `CoursePhaseMailing` component.

### 5. Ensure Proper Routing

- Register the `MailingPage` component in the application's **routes** and **sidebar** configuration to make it accessible for the appropriate course phase.

### 6. Ensure Styling

Nothing to do. `CoursePhaseMailing` renders a `minimal-tiptap` rich-text editor whose stylesheet is
imported by the shared package's own module, and the utilities it needs come from `clients/core`,
which builds the single Tailwind stylesheet for the whole shell and scans every component directory.

Your phase component has no Tailwind entry of its own and must not add one. See the styling section
of `docs/contributor/guide/client.md`.
