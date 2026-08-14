export interface Exercise {
  id: number;
  name: string;
  description?: string;
  imageUrl?: string;
  videoUrl?: string;
}

export interface ExerciseInput {
  name: string;
  description?: string;
  imageUrl?: string;
  videoUrl?: string;
}

export interface SessionExerciseInput {
  exerciseId: number;
  series: number;
  repetitions: number;
  order: number;
}

export interface SessionSummary {
  id: number;
  name: string;
  description?: string;
  exerciseCount: number;
}

export interface SessionExercise {
  exercise: Exercise;
  series: number;
  repetitions: number;
  order: number;
}

export interface SessionDetail {
  id: number;
  name: string;
  description?: string;
  exercises: SessionExercise[];
}

export interface SessionInput {
  name: string;
  description?: string;
  exercises: SessionExerciseInput[];
}

export interface RoutineSessionInput {
  sessionId: number;
  day: number;
}

export interface RoutineSummary {
  id: number;
  name: string;
  description?: string;
}

export interface RoutineSession {
  day: number;
  session: SessionDetail;
}

export interface RoutineDetail extends RoutineSummary {
  sessions: RoutineSession[];
}

export interface RoutineInput {
  name: string;
  description?: string;
  sessions: RoutineSessionInput[];
}

export type FieldErrors = Record<string, string[]>;
