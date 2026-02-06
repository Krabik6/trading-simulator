"use client";

import { useState, useRef, useCallback, type ChangeEvent } from "react";
import { Input } from "@/components/ui/input";

const DOMAINS = [
  "gmail.com",
  "yahoo.com",
  "outlook.com",
  "hotmail.com",
  "icloud.com",
  "mail.ru",
  "yandex.ru",
  "proton.me",
];

interface EmailInputProps
  extends Omit<React.ComponentProps<typeof Input>, "onChange" | "value"> {
  value: string;
  onChange: (value: string) => void;
}

export function EmailInput({ value, onChange, ...props }: EmailInputProps) {
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const wrapperRef = useRef<HTMLDivElement>(null);

  const atIndex = value.indexOf("@");
  const localPart = atIndex >= 0 ? value.slice(0, atIndex) : value;
  const typedDomain = atIndex >= 0 ? value.slice(atIndex + 1) : "";

  const suggestions =
    atIndex >= 0 && localPart.length > 0
      ? DOMAINS.filter(
          (d) => d.startsWith(typedDomain) && d !== typedDomain,
        ).slice(0, 5)
      : [];

  const showSuggestions = open && suggestions.length > 0;

  const pick = useCallback(
    (domain: string) => {
      onChange(`${localPart}@${domain}`);
      setOpen(false);
      setActiveIndex(-1);
    },
    [localPart, onChange],
  );

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value);
    setOpen(true);
    setActiveIndex(-1);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!showSuggestions) return;

    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => (i + 1) % suggestions.length);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => (i <= 0 ? suggestions.length - 1 : i - 1));
    } else if (e.key === "Enter" && activeIndex >= 0) {
      e.preventDefault();
      pick(suggestions[activeIndex]);
    } else if (e.key === "Escape") {
      setOpen(false);
    } else if (e.key === "Tab" && suggestions.length > 0) {
      e.preventDefault();
      pick(suggestions[activeIndex >= 0 ? activeIndex : 0]);
    }
  };

  return (
    <div ref={wrapperRef} className="relative">
      <Input
        type="email"
        autoComplete="email"
        value={value}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        onFocus={() => setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
        {...props}
      />
      {showSuggestions && (
        <ul className="border-border bg-popover text-popover-foreground absolute top-full z-50 mt-1 w-full overflow-hidden rounded-md border shadow-md">
          {suggestions.map((domain, i) => (
            <li
              key={domain}
              className={`cursor-pointer px-3 py-1.5 text-sm ${
                i === activeIndex ? "bg-accent text-accent-foreground" : ""
              }`}
              onMouseDown={(e) => {
                e.preventDefault();
                pick(domain);
              }}
              onMouseEnter={() => setActiveIndex(i)}
            >
              {localPart}@<span className="text-muted-foreground">{domain}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
