---
sidebar_position: 4
---

# 📊 Assessment Course Phase

The **Assessment Phase** allows instructors to evaluate students across multiple competencies using a structured scoring system. Students can be assessed through self-evaluation, peer evaluation, and tutor/instructor evaluation. Students may also evaluate their tutors and team leads.

---

## 📝 Using the Assessment Phase

Once the assessment phase is set up, students and instructors can begin evaluations. Here's how to navigate and complete assessments effectively.

### For Students

1. **Access the Phase**: Log in and navigate to the assessment phase in your course dashboard.
2. **Self-Evaluation**: If enabled, rate your own competencies using the provided template.
3. **Peer Evaluation**: Evaluate team members based on the peer template (if applicable).
4. **Tutor Evaluation**: Provide feedback on tutors or leads using the designated template.
5. **Submit**: Ensure all evaluations are submitted and _marked as final_ before the deadline.

### For Instructors/Tutors

1. **Review Submissions**: Monitor student progress and submissions in the phase overview.
2. **Conduct Assessments**: Use the tutor template to assess students.
3. **Provide Feedback**: Offer constructive comments to support student development.

> 💡 **Tip**: Regularly check for updates and communicate with your team to ensure timely submissions.

This section guides you through the practical use of the assessment phase for a smooth evaluation process.

📺 **Video Tutorial**: [Watch here!](https://live.rbg.tum.de/w/artemisintro/62136)

## ⚙️ Setting Up the Assessment Phase

To configure the assessment phase, follow these steps:

### 1. Add an Assessment Phase

Use the Course Configurator to add an **Assessment Phase** to your course structure. Connect the participants and team data (if you want to use peer or tutor evaluations).

### 2. Open the Configuration

In the course sidebar, select the newly added assessment phase, then click on: `Settings` (gear icon).

This opens a dialog for setting the phase's behavior and schedule.

### 3. Configure Assessment Phase Settings

You must define the following core settings:

| Setting                 | Description                                                                |
| ----------------------- | -------------------------------------------------------------------------- |
| **Assessment Template** | Select which assessment template to use (defines the competency structure) |
| **Start Date**          | When the assessment phase begins (using the timeframe selector)            |
| **Deadline**            | When the assessment phase ends (using the timeframe selector)              |

#### Visibility Settings

Control when students can see their assessment results:

| Setting                       | Description                                                                                       |
| ----------------------------- | ------------------------------------------------------------------------------------------------- |
| **Grade Suggestions Visible** | When enabled, students can view their grade suggestions after the deadline has passed             |
| **Action Items Visible**      | When enabled, students can see their action items (feedback points) after the deadline has passed |

> 💡 **Note**: Both visibility settings only take effect **after the assessment deadline**. Before the deadline, grade suggestions and action items are never visible to students, regardless of these settings. This ensures students cannot see preliminary results during the active assessment phase.

**Use Cases:**

- **Grade Suggestions**: Enable this when you want students to see their overall performance score
- **Action Items**: Enable this to share specific feedback points and improvement suggestions with students
- **Both Disabled**: Use when you want to keep all assessment results private (e.g., for administrative purposes only)

#### Optional Evaluation Types

You can enable additional evaluation methods

| Evaluation Type      | Description                                                        |
| -------------------- | ------------------------------------------------------------------ |
| **Self Evaluation**  | Students evaluate themselves using a separate competency template  |
| **Peer Evaluation**  | Students evaluate their peers using a separate competency template |
| **Tutor Evaluation** | Students evaluate tutors using a separate competency template      |

Each evaluation type has its own:

- **Template**: Which competencies to evaluate
- **Start Date**: When this evaluation becomes available
- **Deadline**: When this evaluation closes

> ⚠️ **Note**: All timestamps use **system time (Europe/Berlin) with deadlines ending at 23:59**.

---

## 🏗️ Creating and Editing Assessment Schemas

Assessment schemas define the structure of your evaluation through **Assessment Templates**.

### Understanding the Hierarchy

The assessment system uses a three-level hierarchy:

```text
Assessment Template
├── Categories (weighted groups)
│   ├── Competency 1 (individual skills with weights)
│   ├── Competency 2
│   └── ...
└── ...
```

### Creating a New Assessment Template

1. **Navigate to Settings**: In your assessment phase, go to the Settings page
2. **Create Template**: Click "Create New Template" in the template selector
3. **Fill Template Details**:
   - **Name**: A descriptive name for your template
   - **Description**: Brief explanation of the template's purpose

### Building the Competency Structure

Once your template is created, you need to build the competency structure:

#### Adding Categories

Categories are weighted groups of related competencies:

1. Click **"Add Category"**
2. Fill in the category details:
   - **Name**: Category title (e.g., "Technical Skills")
   - **Short Name**: Abbreviated version
   - **Description**: Brief explanation
   - **Weight**: Numerical weight for final score calculation
3. Click **"Create"**

#### Adding Competencies

Competencies are individual skills within categories:

1. Click **"Add Competency"** within a category
2. Complete the competency form:
   - **Name**: Competency title (e.g., "Programming Ability")
   - **Short Name**: Abbreviated version
   - **Weight**: Numerical weight within the category
   - **Description**: Brief explanation
   - **Score Level Descriptions**: Define what each level means:
     - **Very Bad**: Description for lowest performance
     - **Bad**: Description for below-average performance
     - **Ok**: Description for average performance
     - **Good**: Description for above-average performance
     - **Very Good**: Description for excellent performance

### Editing Existing Templates

- **Edit Categories**: Click the edit icon next to any category
- **Edit Competencies**: Click the edit icon next to any competency
- **Delete Items**: Click the trash icon (⚠️ this cannot be undone)

📺 **Video Tutorial**: Coming soon

---

## 🤝 Showing Self and Peer Evaluation Results

Self and peer evaluation results appear in the assessment view automatically when the self/peer evaluation uses the **same schema** as the assessment. Because both then share the same competencies, each evaluation score lines up with its assessment competency without any extra configuration.

If a **different schema** is configured for self or peer evaluation, those scores are not shown in the assessment view (feedback items are still shown).

📺 **Video Tutorial**: Coming soon

---

## 📈 How Average Scores Are Calculated

The assessment system uses a sophisticated weighted scoring algorithm to compute final scores.

### Score Level Conversion

Each score level is converted to a numerical value:

| Score Level | Numerical Value |
| ----------- | --------------- |
| Very Good   | 1.0             |
| Good        | 2.0             |
| Ok          | 3.0             |
| Bad         | 4.0             |
| Very Bad    | 5.0             |

### Calculation Process

The `getWeightedScoreLevel` function computes the final score through these steps:

#### 1. **Competency-Level Calculation**

For each competency with scores:

```text
Competency Score = (Sum of all scores × competency weight) / number of scores / total category competency weights
```

#### 2. **Category-Level Aggregation**

For each category:

```text
Category Score = Sum of all competency scores in category
Weighted Category Score = Category Score × category weight
```

#### 3. **Final Score Calculation**

```text
Final Score = Sum of all weighted category scores / Sum of all category weights
```

### Key Features

- **Weight-Based**: Both competency and category weights influence the final score
- **Proportional**: Only categories and competencies with actual scores contribute to the calculation
- **Flexible**: Handles partial assessments (not all competencies need scores)
- **Normalized**: Final score is normalized by the total weights of assessed categories

---

## 🎯 Best Practices

### Template Design

- **Clear Descriptions**: Write specific, measurable descriptions for each score level
- **Balanced Weights**: Distribute weights logically across categories and competencies
- **Consistent Granularity**: Keep similar levels of detail across all competencies

### Self and Peer Evaluation

- **Shared Schema**: Use the same schema for assessment and self/peer evaluation so evaluation results show up in the assessment view
- **Consistency**: Keep competency descriptions identical across the shared schema so scores stay comparable

---

### Getting Help

If you encounter issues not covered here:

1. Check that all required fields are filled
2. Verify date ranges and permissions
3. Contact your system administrator for technical support

---

✅ **You're all set!** Your assessment phase is now configured and ready for student evaluations.
