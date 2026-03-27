---
sidebar_position: 4
---

# How to Add a New Microfrontend

This guide explains how to add a new microfrontend component (ending with `_component`) to this repository, using our existing template structure.

---

## 1. Create Your New Microfrontend Inside This Repo

### 1.1. Copy the Template Microfrontend

1. **Naming**

   - The name of your new microfrontend **must** end with `_component`.
   - Example:
     - **Correct**: `yourAmazingMicrofrontend_component`
     - **Incorrect**: `yourAmazingMicrofrontend`
   - **Reason**: The TypeScript declarations in `core/src/declaration.d.ts` recognize names ending in `*_component`. Using any other naming convention will cause type errors.

2. **Duplicate the Template**

   - Copy the entire `template_component` folder.
   - Rename it to your new component’s name, ensuring you include the `_component` suffix.

3. **Update `webpack.config.mjs`**

   - In your new component’s `webpack.config.mjs`, update the following constants:
     - `COMPONENT_NAME` (set this to your component’s name)
     - `COMPONENT_DEV_PORT` (assign a unique port for local development, e.g. `3001`. Make sure it does not collide with other components in this repo.)

4. **Adjust the Naming in Other Files**

   - **`<yourName_component>/Dockerfile`**: Rename all instances of `template_component` to your new component name.
   - **`<yourName_component>/package.json`**: Update the `"name"` field to match your component.

5. **Register Your Microfrontend in the Monorepo**

   - **`lerna.json`** > `packages`
   - **`package.json`** > `packages`
   - **`eslint.config.mjs`** > `workspaceFolders`

   Make sure your new component folder is listed so it’s recognized by the build and lint processes.

6. **Install Dependencies**

   - Run `yarn install` at the repository root to ensure your new microfrontend is properly integrated.

7. **Implement Your Functionality**

   - Add pages, components, and other logic as needed for your new microfrontend.

8. **Export Your Functionality via Routes and Sidebar**

   - **Configure your sidebar** in `sidebar/index.tsx`

     - You can define a single sidebar item that contains multiple subitems.
     - Top-level sidebar items can display an icon (subitems do not).
     - Both items and subitems can have different permission restrictions.
     - **Important:** The top-level sidebar item should have the lowest required permissions. Subitems are only visible if their parent item is accessible.

   - **Configure your routes** in `routes/index.tsx`
     - Ensure the routes match the sidebar structure and permission settings.
     - Each sidebar path must correspond to a route with the same access restrictions.
     - You may add extra routes that are not listed in the sidebar, but do not create sidebar entries without corresponding routes.

---

## 2. Integrate Your Microfrontend with the Core

### 2.1. Update `core/webpack.config.mjs`

1. Determine the URL of your new microfrontend (e.g., via an environment variable).
2. In the `ModuleFederationPlugin > remotes` section, add an entry for your new microfrontend so it can be dynamically loaded.

### 2.2. Set Up External Routes and Sidebars

IMPORTANT: Make sure to exactly follow and copy the files such that the permission control is correctly applied.

1. **Routes**

   - Copy `TemplateRoutes.tsx` (located in `core/src/managementConsole/PhaseMapping/ExternalRoutes`) and rename it for your component.
   - Update its import paths to point to your new component.

2. **Sidebar**

   - Copy `TemplateSidebar.tsx` (located in `core/src/managementConsole/PhaseMapping/ExternalSidebars`) and rename it for your component.
   - Update its import paths accordingly.

3. **Why Static Imports? Why Copying the import logic**
   - Yes! it would be possible to reduce code duplication by introducing a separate class to abstract the import logic.
   - Yes! We have already implemented that but decided to discard it again.
   - Why?
     - Webpack requires static imports for code splitting and analysis. Hence, import paths need to be static and not in form of a variable. This could be resolved by a dictionary, but this would introduce another part which needs to be adjusted
     - Main reason: Dynamically passed import path (i.e. through a dictionary) can introduce repeated reloading and undesired "Loading" states. Our current approach ensures components are loaded and cached effectively. The downside of code repetition of the loading logic seems neglectable (for now.)

### 2.3. Map the Microfrontend to Course Phases

We assume that every component is associated to one course phase type.
The course phase types are stored in the DB. We assume here that you already have a course phase type to which you want to map your component.

1. **Sidebar Mapping**

   - In `core/src/managementConsole/PhaseMapping/PhaseSidebarMapping.tsx`, map the course phase type name to your new `<YourName>Sidebar`.

2. **Router Mapping**
   - In `core/src/managementConsole/PhaseMapping/PhaseRouterMapping.tsx`, map the same course phase type name to your new `<YourName>Routes`.

---

## 3. Deploy Your Microfrontend

1. **Environment Variables**

   - If your component references an environment variable (e.g., `COMPONENT_SUBPATH`), add it in GitHub’s settings and add it in `core/public/env.template.js` so that it is available at runtime.

2. **GitHub Workflows**

   - **`build-and-push-clients.yml`**: Add another output image tag and the building script for for your component. Use the template component as example.
   - **`deploy-docker`**: Include the new image tag (and path env variable) in the “SSH to VM and create .env.prod file” step to ensure Docker Compose picks it up.
   - **`dev.yml`**: Reference the new image tag.
   - **`prod.yml`**: Reference the new image tag.

3. **Docker Compose**
   - In `docker-compose.prod.yml`, follow the existing template for deployment. Rename or adjust the middlewares/services as needed to match your new component.

---

## 2. Create a New Microfrontend Outside This Repo

> **Not yet supported.**

**Note:**

- For an external component, the `shared_library` must be published in full, and you must adjust the `publicPath` accordingly in `webpack.config.mjs`.

---

**Happy coding!** If you run into issues or have questions, please reach out @niclasheun or refer to existing components for reference.
