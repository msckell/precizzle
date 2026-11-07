import { useId } from "react";

export function Decorations() {
  const id = useId().replace(/[^\w-]/g, "");
  const url = (name: string) => `url(#${id}-${name})`;
  return (
    <div className="decorations" aria-hidden="true">
      <svg
        className="decoration bottom-left"
        viewBox="0 0 400 300"
        preserveAspectRatio="none"
      >
        <defs>
          <linearGradient id={`${id}-bl`} x1="0" y1="1" x2="1" y2="0">
            <stop offset="0" stopColor="#84EDCE" stopOpacity=".5" />
            <stop offset=".55" stopColor="#BDF0D8" stopOpacity=".3" />
            <stop offset="1" stopColor="#BDF0D8" stopOpacity="0" />
          </linearGradient>
          <linearGradient id={`${id}-bl2`} x1="0" y1="1" x2=".8" y2="0">
            <stop offset="0" stopColor="#3ED9AB" stopOpacity=".2" />
            <stop offset="1" stopColor="#3ED9AB" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path
          d="M0 300 L0 40 C 70 40, 130 130, 200 195 C 260 250, 330 280, 400 300 Z"
          fill={url("bl")}
        />
        <path
          d="M0 300 L0 140 C 80 140, 150 215, 240 258 C 290 282, 340 296, 370 300 Z"
          fill={url("bl2")}
        />
      </svg>
      <svg
        className="decoration top-right"
        viewBox="0 0 300 180"
        preserveAspectRatio="none"
      >
        <defs>
          <linearGradient id={`${id}-tr`} x1="1" y1="0" x2="0" y2=".7">
            <stop offset="0" stopColor="#FC7A78" stopOpacity=".3" />
            <stop offset=".4" stopColor="#FF7925" stopOpacity=".16" />
            <stop offset=".75" stopColor="#FDBF43" stopOpacity=".08" />
            <stop offset="1" stopColor="#FDBF43" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path
          d="M300 0 L20 0 C 80 20, 120 85, 190 108 C 235 123, 272 148, 300 180 Z"
          fill={url("tr")}
        />
        <path
          d="M300 0 L150 0 C 190 25, 230 55, 300 70 Z"
          fill="#FC7A78"
          fillOpacity=".08"
        />
      </svg>
      <svg
        className="decoration bottom"
        viewBox="0 0 300 80"
        preserveAspectRatio="none"
      >
        <defs>
          <linearGradient id={`${id}-br`} x1="1" y1="1" x2="0" y2="0">
            <stop offset="0" stopColor="#FF7925" stopOpacity=".12" />
            <stop offset=".5" stopColor="#FDBF43" stopOpacity=".09" />
            <stop offset="1" stopColor="#FDBF43" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d="M0 80 C 70 45, 170 18, 300 22 L300 80 Z" fill={url("br")} />
      </svg>
    </div>
  );
}
