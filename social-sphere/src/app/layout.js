import "./globals.css";
import { ThemeProvider } from "next-themes";
import { ThemeToggle } from "@/components/layout/ThemeToggle";

export const metadata = {
  title: "SocialSphere",
  description: "SocialSphere",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          {children}
          <ThemeToggle />
        </ThemeProvider>
      </body>
    </html>
  );
}
