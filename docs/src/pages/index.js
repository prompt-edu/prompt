import Link from "@docusaurus/Link";
import Layout from "@theme/Layout";
import Heading from "@theme/Heading";
import styles from "./index.module.css";
import {
  Layers,
  FileUser,
  UserCheck,
  Mic,
  Upload,
  Users,
  Clipboard,
  Bug,
  Lightbulb,
  Package,
  Plus,
} from "lucide-react";
import useBaseUrl from "@docusaurus/useBaseUrl";

export default function Home() {
  const promptLogo = useBaseUrl("/img/prompt_logo.svg");
  return (
    <Layout
      title="PROMPT - Course Management Platform"
      description="Project-Oriented Modular Platform for Teaching. A flexible course management tool for project-based university courses."
    >
      <div className="container margin-vert--lg">
        {/* Hero Section */}
        <div className={styles.heroSection}>
          <div className={styles.heroTitleContainer}>
            <img
              src={promptLogo}
              alt="PROMPT Logo"
              className={styles.heroLogo}
              width={100}
              height={100}
            />
            <Heading as="h1" className={styles.heroTitle}>
              PROMPT
            </Heading>
          </div>
          <p className={styles.heroSubtitle}>
            Project-Oriented Modular Platform for Teaching
          </p>
          <p className={styles.heroDescription}>
            A flexible course management tool designed for project-based
            university courses. Streamline administrative tasks and enhance the
            learning experience for both students and instructors.
          </p>
        </div>

        {/* Core Features */}
        <section className={styles.section}>
          <Heading as="h2" className={styles.sectionTitle}>
            Core Features
          </Heading>
          <p className={styles.sectionDescription}>
            The core features are built-in functionalities essential for course
            management, while dynamically loaded phases are additional,
            customizable components that can be added as needed.
          </p>
          <div className={styles.featureGrid}>
            <a
              className={styles.featureCard}
              href="/prompt2/user/course_configurator"
            >
              <div className={styles.featureIcon}>
                <Layers size={26} />
              </div>
              <h2>Course Configuration</h2>
              <p>
                Build a course by assembling various course phases to suit your
                specific teaching needs.
              </p>
            </a>
            <a className={styles.featureCard} href="/prompt2/user/application">
              <div className={styles.featureIcon}>
                <FileUser size={26} />
              </div>
              <h2>Application Phase</h2>
              <p>
                Streamline the application process for courses, making it easier
                for students to apply and for instructors to manage
                applications.
              </p>
            </a>
            <a
              className={styles.featureCard}
              href="/prompt2/user/course_configurator"
            >
              <div className={styles.featureIcon}>
                <UserCheck size={26} />
              </div>
              <h2>Student Management</h2>
              <p>
                Efficiently manage student information and course participation.
              </p>
            </a>
          </div>
        </section>

        {/* Dynamic Phases */}
        <section className={styles.section} style={{ marginTop: "100px" }}>
          <Heading as="h2" className={styles.sectionTitle}>
            Dynamic Course Phases
          </Heading>
          <p className={styles.sectionDescription}>
            PROMPT allows instructors to create and manage independent course
            phases, fostering a collaborative and easily extensible platform for
            project-based learning.
          </p>
          <div className={styles.featureGrid}>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Mic size={26} />
              </div>
              <h2>Interview Phase</h2>
              <p>
                Manage and schedule interviews with applicants as part of the
                selection process.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Upload size={26} />
              </div>
              <h2>TUM Matching Export</h2>
              <p>
                Export data in a format compatible with TUM Matching for
                seamless integration.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Users size={26} />
              </div>
              <h2>Team Phase</h2>
              <p>
                Assign students to teams and projects, and manage project work
                throughout the course.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Clipboard size={26} />
              </div>
              <h2>Assessment Phase</h2>
              <p>
                Conduct structured peer, self, and instructor assessments using
                a configurable framework.
              </p>
            </div>
            <div className={styles.featureCardCustomPhase}>
              <div className={styles.featureIcon}>
                <Plus size={26} />
              </div>
              <h2>Custom Course Phase</h2>
              <p>
                Easily extend PROMPT with custom phases tailored to your course
                needs.
              </p>
            </div>
          </div>
        </section>

        {/* Quick Links */}
        <section className={styles.section} style={{ marginTop: "100px" }}>
          <Heading as="h2" className={styles.sectionTitle}>
            Get in Touch
          </Heading>
          <div className={styles.featureGrid}>
            <div className={styles.linkCard}>
              <div className={styles.featureIcon}>
                <Bug size={26} />
              </div>
              <h3>Report a Bug</h3>
              <p>
                <a
                  href="https://github.com/prompt-edu/prompt/issues"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Report an issue →
                </a>
              </p>
            </div>
            <div className={styles.linkCard}>
              <div className={styles.featureIcon}>
                <Lightbulb size={26} />
              </div>
              <h3>Request a Feature</h3>
              <p>
                <a
                  href="https://github.com/prompt-edu/prompt/issues"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Suggest a feature →
                </a>
              </p>
            </div>
            <div className={styles.linkCard}>
              <div className={styles.featureIcon}>
                <Package size={26} />
              </div>
              <h3>Release Notes</h3>
              <p>
                <a
                  href="https://github.com/prompt-edu/prompt/releases"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  View releases →
                </a>
              </p>
            </div>
          </div>
        </section>
      </div>
    </Layout>
  );
}
