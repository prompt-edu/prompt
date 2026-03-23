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
  Github,
  BookOpen,
  ExternalLink,
  PlayCircle,
} from "lucide-react";
import useBaseUrl from "@docusaurus/useBaseUrl";

const DEMOS = [
  {
    id: "course-configurator",
    label: "Course Configurator",
    description:
      "Build your course by assembling phases via an intuitive drag-and-drop interface.",
    gif: "/img/gifs/course-configurator.gif",
    available: true,
  },
  {
    id: "application",
    label: "Application Phase",
    description:
      "Students apply to courses through a streamlined form — instructors review, filter, and accept applicants.",
    gif: null,
    available: false,
  },
  {
    id: "interview",
    label: "Interview Scheduling",
    description:
      "Coordinate and schedule interviews with applicants as part of the selection process.",
    gif: null,
    available: false,
  },
  {
    id: "team-allocation",
    label: "Team Allocation",
    description:
      "Assign students to teams and projects, manually or through automated matching.",
    gif: null,
    available: false,
  },
  {
    id: "assessment",
    label: "Assessment",
    description:
      "Run structured peer, self, and instructor assessments using a configurable rubric framework.",
    gif: null,
    available: false,
  },
];

function DemoShowcase() {
  const firstAvailable = DEMOS.find((d) => d.available);

  return (
    <section className={styles.section}>
      <Heading as="h2" className={styles.sectionTitle}>
        See PROMPT in Action
      </Heading>
      <p className={styles.sectionDescription}>
        Watch how PROMPT simplifies course management — from setup to grading.
      </p>

      {firstAvailable && (
        <div className={styles.demoFeatured}>
          <div className={styles.demoFeaturedContent}>
            <div className={styles.demoFeaturedMeta}>
              <span className={styles.demoBadge}>
                <PlayCircle size={14} />
                {firstAvailable.label}
              </span>
              <p className={styles.demoFeaturedDescription}>
                {firstAvailable.description}
              </p>
            </div>
            <div className={styles.demoGifWrapper}>
              <img
                src={useBaseUrl(firstAvailable.gif)}
                alt={`${firstAvailable.label} demo`}
                className={styles.demoGif}
              />
            </div>
          </div>
        </div>
      )}

      <div className={styles.demoGrid}>
        {DEMOS.filter((d) => !d.available).map((demo) => (
          <div key={demo.id} className={styles.demoCard}>
            <div className={styles.demoCardHeader}>
              <span className={styles.demoComingSoonBadge}>Coming soon</span>
              <h3 className={styles.demoCardTitle}>{demo.label}</h3>
            </div>
            <p className={styles.demoCardDescription}>{demo.description}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

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
            A flexible course management platform designed for project-based
            university courses. Streamline administrative tasks and enhance the
            learning experience for both students and instructors.
          </p>
          <div className={styles.heroActions}>
            <a
              href="https://prompt.aet.cit.tum.de/"
              target="_blank"
              rel="noopener noreferrer"
              className={styles.heroCTAPrimary}
            >
              <ExternalLink size={16} />
              Live Demo
            </a>
            <Link to="/user/course_configurator" className={styles.heroCTASecondary}>
              <BookOpen size={16} />
              Documentation
            </Link>
            <a
              href="https://github.com/prompt-edu/prompt"
              target="_blank"
              rel="noopener noreferrer"
              className={styles.heroCTASecondary}
            >
              <Github size={16} />
              GitHub
            </a>
          </div>
        </div>

        {/* Demo Showcase */}
        <DemoShowcase />

        {/* Core Features */}
        <section className={styles.section} style={{ marginTop: "80px" }}>
          <Heading as="h2" className={styles.sectionTitle}>
            Core Features
          </Heading>
          <p className={styles.sectionDescription}>
            Built-in functionalities essential for course management, with
            dynamically loaded phases that extend the platform as needed.
          </p>
          <div className={styles.featureGrid}>
            <a
              className={styles.featureCard}
              href="/prompt/user/course_configurator"
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
            <a className={styles.featureCard} href="/prompt/user/application">
              <div className={styles.featureIcon}>
                <FileUser size={26} />
              </div>
              <h2>Application Phase</h2>
              <p>
                Streamline the application process for courses, making it easier
                for students to apply and for instructors to manage applications.
              </p>
            </a>
            <a
              className={styles.featureCard}
              href="/prompt/user/course_configurator"
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
        <section className={styles.section} style={{ marginTop: "80px" }}>
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

        {/* Get in Touch */}
        <section className={styles.section} style={{ marginTop: "80px" }}>
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
