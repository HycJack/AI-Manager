import type { Preferences } from "./models";

export declare function LoadPrefs(): Promise<Preferences>;
export declare function SavePrefs(prefs: Preferences): Promise<void>;
