import { useState, useEffect, useRef } from "react";
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
  ChevronDown,
} from "lucide-react";
import useBaseUrl from "@docusaurus/useBaseUrl";

const showcaseItems = [
  {
    title: "Course Configuration",
    description:
      "Design your course by assembling independent phases with a drag-and-drop configurator. Link outputs between phases, configure participation settings, and tailor the entire course structure to your exact teaching needs — no technical knowledge required.",
    gifSrc: "/img/gifs/course-configurator.gif",
    gifAlt: "Course Configurator demo",
    link: "/user/course_configurator",
    linkLabel: "Learn about course configuration →",
    tag: "Core Feature",
  },
  {
    title: "Application & Student Management",
    description:
      "Streamline the entire application lifecycle. Students apply online, instructors review and manage submissions, and accepted participants are automatically onboarded into the course.",
    gifSrc: "/img/gifs/application-management.gif",
    gifAlt: "Application management demo",
    link: "/user/application",
    linkLabel: "Learn about applications →",
    tag: "Core Feature",
    reversed: true,
  },
  {
    title: "Assessment",
    description:
      "PROMPT's assessment engine is one of its most powerful capabilities. Configure multi-criteria rubrics, run structured peer reviews, self-assessments, and instructor evaluations — all in one unified workflow. Results are aggregated automatically, giving instructors instant, detailed insight into student performance across the entire cohort.",
    gifSrc: "/img/gifs/assessment.gif",
    gifAlt: "Assessment demo",
    link: "/user/assessment",
    linkLabel: "Learn about assessments →",
    tag: "Highlight",
  },
];

function ShowcaseSlide({ item, active }) {
  const gifSrc = useBaseUrl(item.gifSrc || "");
  return (
    <div
      className={`${styles.showcaseSlide} ${active ? styles.showcaseSlideActive : ""} ${item.reversed ? styles.showcaseSlideReversed : ""}`}
      aria-hidden={!active}
    >
      <div className={styles.showcaseMedia}>
        {item.gifSrc ? (
          <img src={gifSrc} alt={item.gifAlt} className={styles.showcaseGif} />
        ) : (
          <div className={styles.gifPlaceholder}>
            <span className={styles.gifPlaceholderText}>Demo coming soon</span>
          </div>
        )}
      </div>
      <div className={styles.showcaseText}>
        {item.tag && (
          <span
            className={`${styles.showcaseTag} ${item.tag === "Highlight" ? styles.showcaseTagHighlight : ""}`}
          >
            {item.tag}
          </span>
        )}
        <h3 className={styles.showcaseTitle}>{item.title}</h3>
        <p className={styles.showcaseDescription}>{item.description}</p>
        <Link to={item.link} className={styles.showcaseLink}>
          {item.linkLabel}
        </Link>
      </div>
    </div>
  );
}

function ShowcaseSection({ items }) {
  const [activeIndex, setActiveIndex] = useState(0);
  const [hasScrolled, setHasScrolled] = useState(false);
  const containerRef = useRef(null);
  const rafRef = useRef(null);

  useEffect(() => {
    const compute = () => {
      if (!containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const totalScrollable = rect.height - window.innerHeight;
      if (totalScrollable <= 0) return;
      const scrolled = -rect.top;
      if (scrolled > 60) setHasScrolled(true);
      const progress = Math.max(0, Math.min(1 - 1e-9, scrolled / totalScrollable));
      const newIndex = Math.min(
        Math.floor(progress * items.length),
        items.length - 1
      );
      setActiveIndex(Math.max(0, newIndex));
    };

    const handleScroll = () => {
      if (rafRef.current) return;
      rafRef.current = requestAnimationFrame(() => {
        compute();
        rafRef.current = null;
      });
    };

    window.addEventListener("scroll", handleScroll, { passive: true });
    compute();
    return () => {
      window.removeEventListener("scroll", handleScroll);
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, [items.length]);

  return (
    <div
      ref={containerRef}
      className={styles.showcaseScrollContainer}
      style={{ height: `${items.length * 100}vh` }}
    >
      <div className={styles.showcaseSticky}>
        {items.map((item, i) => (
          <ShowcaseSlide key={i} item={item} active={i === activeIndex} />
        ))}

        {/* Progress dots */}
        <div className={styles.showcaseProgress}>
          {items.map((it, i) => (
            <div
              key={i}
              className={`${styles.progressDot} ${i === activeIndex ? styles.progressDotActive : ""}`}
              title={it.title}
            />
          ))}
        </div>

        {/* Step counter */}
        <div className={styles.showcaseCounter}>
          <span className={styles.showcaseCounterActive}>
            {String(activeIndex + 1).padStart(2, "0")}
          </span>
          <span className={styles.showcaseCounterSep}> / </span>
          <span>{String(items.length).padStart(2, "0")}</span>
        </div>

        {/* Scroll hint */}
        <div
          className={`${styles.scrollHint} ${hasScrolled ? styles.scrollHintHidden : ""}`}
        >
          <ChevronDown size={18} />
          <span>Scroll to explore</span>
        </div>
      </div>
    </div>
  );
}

export default function Home() {
  const promptLogo = useBaseUrl("/img/prompt_logo.svg");
  return (
    <Layout
      title="PROMPT - Course Management Platform"
      description="Project-Oriented Modular Platform for Teaching. A flexible course management tool for project-based university courses."
    >
      {/* Hero */}
      <div className="container">
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
            A flexible, open-source course management platform built for
            project-based university courses. Streamline every phase — from
            applications to assessments — in one place.
          </p>
          <div className={styles.heroButtons}>
            <Link className={styles.heroButtonPrimary} to="/user">
              Get Started
            </Link>
            <Link
              className={styles.heroButtonSecondary}
              href="https://github.com/prompt-edu/prompt"
              target="_blank"
              rel="noopener noreferrer"
            >
              View on GitHub
            </Link>
          </div>
        </div>
      </div>

      {/* See it in Action — full-bleed showcase */}
      <section className={styles.showcaseSection}>
        <div className={styles.showcaseSectionHeader}>
          <Heading as="h2" className={styles.sectionTitle}>
            See PROMPT in Action
          </Heading>
          <p className={styles.sectionDescription}>
            Everything you need to run a project-based course — from first
            application to final assessment.
          </p>
        </div>
        <ShowcaseSection items={showcaseItems} />
      </section>

      <div className="container">
        {/* Core Features */}
        <section className={styles.section}>
          <Heading as="h2" className={styles.sectionTitle}>
            Core Features
          </Heading>
          <p className={styles.sectionDescription}>
            Built-in functionalities essential for every course, complemented
            by dynamically loaded phases you can add as needed.
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
                Streamline the application process — easier for students to
                apply, simpler for instructors to manage.
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
                Efficiently manage student information and course participation
                throughout the lifecycle.
              </p>
            </a>
          </div>
        </section>

        {/* Dynamic Phases */}
        <section className={styles.section}>
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
        <section className={styles.section}>
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
