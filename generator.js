import { NotDiamond } from "notdiamond";
import dotenv from "dotenv";
import fs from "fs";
import { execSync } from "child_process";

dotenv.config();

// Constants
const notDiamond = new NotDiamond({
  apiKey: process.env.NOTDIAMOND_API_KEY,
});
const REPO_PATH = "/path/to/your/repo";
const REPO_URL = "git@github.com:your-username/GoRoutineChronicles.git";

// Initialize Git repository if not already initialized
const initializeGitRepo = () => {
  try {
    execSync(`git -C ${REPO_PATH} status`, { stdio: "ignore" });
  } catch {
    console.log("Initializing Git repository...");
    execSync(`mkdir -p ${REPO_PATH}`);
    execSync(`git init`, { cwd: REPO_PATH });
    execSync(`git remote add origin ${REPO_URL}`, { cwd: REPO_PATH });
  }
};

// Request LLM to generate Go code
const generateGoCode = async () => {
  const prompt = `Create a simple, unique Go project with:
- A meaningful package structure.
- One interesting feature (e.g., API server, CLI tool, or data processing).
- A README with a brief description.`;

  console.log("Requesting Go code generation...");
  const result = await notDiamond.create({
    messages: [{ content: prompt, role: "user" }],
    llmProviders: [
      { provider: "openai", model: "gpt-4-turbo" },
      { provider: "anthropic", model: "claude-3" },
    ],
  });

  if ("detail" in result) {
    console.error("Error during code generation:", result.detail);
    return null;
  }

  console.log("Go code generated successfully.");
  return result.function_output; // Return the generated Go code
};

// Create a new Go project
const createGoProject = async () => {
  const timestamp = Date.now();
  const projectDir = `${REPO_PATH}/project_${timestamp}`;
  fs.mkdirSync(projectDir, { recursive: true });

  const code = await generateGoCode();
  if (!code) return null;

  fs.writeFileSync(`${projectDir}/main.go`, code);

  // Create a README
  fs.writeFileSync(
    `${projectDir}/README.md`,
    `# Auto-Generated Go Project\n\nThis project was generated using an LLM.\n\n## Usage\n\nRun with:\n\`\`\`\ngo run main.go\n\`\`\``
  );

  console.log(`Go project created at ${projectDir}`);
  return projectDir;
};

// Commit and push changes to GitHub
const commitAndPush = (projectDir) => {
  console.log("Committing and pushing changes...");
  execSync(`git add .`, { cwd: REPO_PATH });
  execSync(`git commit -m "Daily Go project - ${new Date().toISOString()}"`, {
    cwd: REPO_PATH,
  });
  execSync(`git push origin main`, { cwd: REPO_PATH });
};

// Main automation process
const main = async () => {
  initializeGitRepo();

  const projectDir = await createGoProject();
  if (projectDir) {
    execSync(`mv ${projectDir}/* ${REPO_PATH}/`, { cwd: REPO_PATH });
    commitAndPush(projectDir);
  }
};

// Run the main function
main().catch((err) => {
  console.error("Error during execution:", err);
});
