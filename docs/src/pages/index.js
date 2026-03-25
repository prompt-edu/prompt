import { useEffect, useRef } from "react";
import Link from "@docusaurus/Link";
import Layout from "@theme/Layout";
import Heading from "@theme/Heading";
import styles from "./index.module.css";
import {
  Layers,
  FileUser,
  UserCheck,
  Mic,
  GitMerge,
  Users,
  UserPlus,
  Clipboard,
  Award,
  Terminal,
  Bug,
  Lightbulb,
  Package,
  Plus,
} from "lucide-react";
import useBaseUrl from "@docusaurus/useBaseUrl";

const SHOWCASE_ITEMS = [
  {
    title: "Course Configuration",
    description:
      "Design your course by assembling independent phases with a drag-and-drop configurator. Link outputs between phases, configure participation settings, and tailor the entire course structure to your exact teaching needs — no technical knowledge required.",
    mp4Src: "/img/gifs/course-configurator.mp4",
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
    mp4Src: "/img/gifs/application-management.mp4",
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
    mp4Src: "/img/gifs/assessment.mp4",
    gifSrc: "/img/gifs/assessment.gif",
    gifAlt: "Assessment demo",
    link: "/user/assessment",
    linkLabel: "Learn about assessments →",
    tag: "Highlight",
  },
];

function ShowcaseItem({ item }) {
  const ref = useRef(null);
  const mp4Src = useBaseUrl(item.mp4Src || "");
  const gifSrc = useBaseUrl(item.gifSrc || "");
  const hasVideo = Boolean(item.mp4Src);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    // Respect the user's motion preference
    if (
      typeof window !== "undefined" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    ) {
      el.style.opacity = "1";
      el.style.transform = "none";
      // Stop any autoplaying media
      const media = el.querySelector("video");
      if (media) {
        media.autoplay = false;
        media.loop = false;
        media.pause();
      }
      return;
    }

    // JS opts into the hidden initial state so no-JS visitors always see content
    el.style.opacity = "0";
    el.style.transform = "translateY(36px)";

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          // Determine entry direction: positive top = coming from below, negative = from above
          const fromBelow = entry.boundingClientRect.top > 0;

          // Snap to starting position with no transition
          el.style.transition = "none";
          el.style.opacity = "0";
          el.style.transform = `translateY(${fromBelow ? "36px" : "-36px"})`;

          // Force reflow so the browser registers the snap before we animate
          void el.offsetHeight;

          // Animate in
          el.style.transition =
            "opacity 0.5s cubic-bezier(0.4,0,0.2,1), transform 0.5s cubic-bezier(0.4,0,0.2,1)";
          el.style.opacity = "1";
          el.style.transform = "translateY(0)";
        } else {
          // Instantly hide — no visible transition since element is off-screen
          el.style.transition = "none";
          el.style.opacity = "0";
          // Pre-position for the next entry direction
          const exitedAbove = entry.boundingClientRect.top < 0;
          el.style.transform = `translateY(${exitedAbove ? "-36px" : "36px"})`;
        }
      },
      { threshold: 0.18 },
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return (
    <div
      ref={ref}
      className={`${styles.showcaseItem} ${item.reversed ? styles.showcaseItemReversed : ""}`}
    >
      <div className={styles.showcaseMedia}>
        {hasVideo ? (
          <video
            className={styles.showcaseGif}
            autoPlay
            loop
            muted
            playsInline
            preload="none"
            aria-label={item.gifAlt}
          >
            <source src={mp4Src} type="video/mp4" />
            {/* GIF fallback for browsers that don't support video */}
            <img
              src={gifSrc}
              alt={item.gifAlt}
              loading="lazy"
              decoding="async"
            />
          </video>
        ) : item.gifSrc ? (
          <img
            src={gifSrc}
            alt={item.gifAlt}
            className={styles.showcaseGif}
            loading="lazy"
            decoding="async"
          />
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
            <a
              className={styles.heroButtonSecondary}
              href="https://github.com/prompt-edu/prompt"
              target="_blank"
              rel="noopener noreferrer"
            >
              View on GitHub
            </a>
          </div>
        </div>
      </div>

      {/* Demos */}
      <section
        className={styles.showcaseSection}
        aria-labelledby="demos-heading"
      >
        <h2 id="demos-heading" className={styles.visuallyHidden}>
          Demos
        </h2>
        {SHOWCASE_ITEMS.map((item, i) => (
          <ShowcaseItem key={i} item={item} />
        ))}
      </section>

      <div className="container">
        {/* Core Features */}
        <section className={styles.section} style={{ marginTop: "80px" }}>
          <Heading as="h2" className={styles.sectionTitle}>
            Core Features
          </Heading>
          <p className={styles.sectionDescription}>
            Built-in functionalities essential for every course, complemented by
            dynamically loaded phases you can add as needed.
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
                Schedule and conduct structured interviews with applicants as
                part of the selection process.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <GitMerge size={26} />
              </div>
              <h2>Matching Phase</h2>
              <p>
                Automatically match students to projects based on application
                scores and preferences using configurable algorithms.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Users size={26} />
              </div>
              <h2>Team Allocation Phase</h2>
              <p>
                Assign students to teams and projects managed by instructors,
                with full control over team composition.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <UserPlus size={26} />
              </div>
              <h2>Self Team Allocation Phase</h2>
              <p>
                Let students form their own teams — they create groups and
                invite peers without instructor involvement.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Clipboard size={26} />
              </div>
              <h2>Assessment Phase</h2>
              <p>
                Conduct structured peer, self, and instructor assessments using
                configurable rubrics and weighted scoring.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Award size={26} />
              </div>
              <h2>Certificate Phase</h2>
              <p>
                Generate and distribute personalized course completion
                certificates using customizable Typst templates.
              </p>
            </div>
            <div className={styles.featureCard}>
              <div className={styles.featureIcon}>
                <Terminal size={26} />
              </div>
              <h2>DevOps Challenge Phase</h2>
              <p>
                Present hands-on DevOps exercises and practical challenges
                to students as part of their coursework.
              </p>
            </div>
            <div className={styles.featureCardCustomPhase}>
              <div className={styles.featureIcon}>
                <Plus size={26} />
              </div>
              <h2>Custom Course Phase</h2>
              <p>
                Easily extend PROMPT with your own phases — the micro-frontend
                architecture makes adding new phases straightforward.
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
