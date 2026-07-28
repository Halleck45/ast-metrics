document.addEventListener("DOMContentLoaded", function () {
  // Intersection Observer for fade-up animations
  const observerOptions = {
    root: null,
    rootMargin: "0px",
    threshold: 0.1
  };

  const observer = new IntersectionObserver((entries, observer) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add("visible");
        observer.unobserve(entry.target);
      }
    });
  }, observerOptions);

  const animatedElements = document.querySelectorAll(".fade-up");
  animatedElements.forEach(el => observer.observe(el));

  // Spotlight Effect for Cards
  const cards = document.querySelectorAll(".feature-card, .ci-card");

  cards.forEach(card => {
    card.addEventListener("mousemove", (e) => {
      const rect = card.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;

      card.style.setProperty("--mouse-x", `${x}px`);
      card.style.setProperty("--mouse-y", `${y}px`);
    });
  });

  // Hero Parallax Effect
  const heroSection = document.querySelector(".hero-section");
  const codeWindow = document.querySelector(".code-window");

  if (heroSection && codeWindow) {
    heroSection.addEventListener("mousemove", (e) => {
      const { clientX, clientY } = e;
      const { innerWidth, innerHeight } = window;

      // Calculate percentage position (-1 to 1)
      const xPos = (clientX / innerWidth - 0.5) * 2;
      const yPos = (clientY / innerHeight - 0.5) * 2;

      // Apply subtle rotation
      // RotateY depends on X position (left/right)
      // RotateX depends on Y position (up/down) - inverted
      const rotateY = xPos * 5; // Max 5 degrees
      const rotateX = -yPos * 5; // Max 5 degrees

      codeWindow.style.transform = `perspective(1000px) rotateY(${rotateY}deg) rotateX(${rotateX}deg)`;
    });

    // Reset on mouse leave
    heroSection.addEventListener("mouseleave", () => {
      codeWindow.style.transform = "perspective(1000px) rotateY(-5deg) rotateX(2deg)";
    });
  }
});
