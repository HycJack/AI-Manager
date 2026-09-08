import type { Release } from "./models";

export declare function CheckForUpdates(): Promise<Release | null>;
export declare function CheckAndInstall(): Promise<void>;
export declare function Restart(): Promise<void>;
export declare function UpdateState(): Promise<string>;
