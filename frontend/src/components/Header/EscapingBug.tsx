import { useState, useRef, useEffect } from 'react';
import { AnimatedBugIcon } from './AnimatedBugIcon';

const ESCAPE_RADIUS = 120;
const CRAWL_SPEED = 1.5;
const DIRECTION_CHANGE_INTERVAL = 2000;
const PAUSE_CHANCE = 0.15;
const PAUSE_DURATION = 800;
const BUG_SIZE = 48;
const MARGIN = 20;

export function EscapingBug() {
  const [isActive, setIsActive] = useState(false);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [rotation, setRotation] = useState(0);

  const bugRef = useRef<HTMLDivElement>(null);
  const startPos = useRef({ x: 0, y: 0 });
  const mousePos = useRef({ x: 0, y: 0 });
  const direction = useRef(Math.random() * Math.PI * 2);
  const isPaused = useRef(false);
  const animationRef = useRef<number | null>(null);

  const handleClick = () => {
    if (!isActive && bugRef.current) {
      // Capture starting position before going fixed
      const rect = bugRef.current.getBoundingClientRect();
      startPos.current = { x: rect.left, y: rect.top };
      setPosition({ x: rect.left, y: rect.top });
      setIsActive(true);
    }
  };

  // Track mouse position
  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      mousePos.current = { x: e.clientX, y: e.clientY };
    };
    window.addEventListener('mousemove', handleMouseMove);
    return () => window.removeEventListener('mousemove', handleMouseMove);
  }, []);

  // Change direction periodically
  useEffect(() => {
    if (!isActive) return;

    const changeDirection = () => {
      if (!isPaused.current) {
        // Random turn: slight adjustment most of the time
        direction.current += (Math.random() - 0.5) * Math.PI * 0.5;

        // Occasionally pause
        if (Math.random() < PAUSE_CHANCE) {
          isPaused.current = true;
          setTimeout(() => {
            isPaused.current = false;
          }, PAUSE_DURATION);
        }
      }
    };

    const interval = setInterval(changeDirection, DIRECTION_CHANGE_INTERVAL);
    return () => clearInterval(interval);
  }, [isActive]);

  // Main animation loop
  useEffect(() => {
    if (!isActive) return;

    const animate = () => {
      const viewportWidth = window.innerWidth;
      const viewportHeight = window.innerHeight;

      setPosition(prev => {
        let newX = prev.x;
        let newY = prev.y;
        let speed = CRAWL_SPEED;

        const bugCenterX = newX + BUG_SIZE / 2;
        const bugCenterY = newY + BUG_SIZE / 2;

        const dx = mousePos.current.x - bugCenterX;
        const dy = mousePos.current.y - bugCenterY;
        const distanceToMouse = Math.sqrt(dx * dx + dy * dy);

        if (distanceToMouse < ESCAPE_RADIUS && distanceToMouse > 0) {
          // Escape from mouse - run away fast!
          speed = CRAWL_SPEED * 3;
          direction.current = Math.atan2(-dy, -dx);
          isPaused.current = false;
        }

        if (!isPaused.current) {
          newX += Math.cos(direction.current) * speed;
          newY += Math.sin(direction.current) * speed;

          // Boundary check - bounce off viewport edges
          if (newX < MARGIN) {
            direction.current = Math.PI - direction.current;
            newX = MARGIN;
          } else if (newX > viewportWidth - BUG_SIZE - MARGIN) {
            direction.current = Math.PI - direction.current;
            newX = viewportWidth - BUG_SIZE - MARGIN;
          }
          if (newY < MARGIN) {
            direction.current = -direction.current;
            newY = MARGIN;
          } else if (newY > viewportHeight - BUG_SIZE - MARGIN) {
            direction.current = -direction.current;
            newY = viewportHeight - BUG_SIZE - MARGIN;
          }
        }

        // Update rotation to face movement direction
        setRotation(direction.current * (180 / Math.PI) + 90);

        return { x: newX, y: newY };
      });

      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [isActive]);

  return (
    <>
      {/* Container that stays in place */}
      <div
        ref={bugRef}
        onClick={!isActive ? handleClick : undefined}
        style={{
          width: BUG_SIZE,
          height: BUG_SIZE,
          cursor: !isActive ? 'pointer' : 'default',
        }}
      >
        {/* Bug only visible here before activation */}
        {!isActive && <AnimatedBugIcon size={BUG_SIZE} isRunning={false} />}
      </div>

      {/* Fixed position bug after activation */}
      {isActive && (
        <div
          style={{
            position: 'fixed',
            left: position.x,
            top: position.y,
            transform: `rotate(${rotation}deg)`,
            zIndex: 9999,
            willChange: 'left, top, transform',
            pointerEvents: 'none',
          }}
        >
          <AnimatedBugIcon size={BUG_SIZE} isRunning={true} />
        </div>
      )}
    </>
  );
}
