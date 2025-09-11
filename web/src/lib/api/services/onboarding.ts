// Onboarding API - Direct functions for onboarding flow

import { BrokleAPIClient } from '../core/client'
import type {
  OnboardingQuestion,
  OnboardingStatus,
  SubmitResponseRequest,
  SubmitResponsesRequest
} from '@/types/onboarding'

// Flexible base client - versions specified per endpoint
const client = new BrokleAPIClient('/api')

// Direct onboarding functions
export const getQuestions = async (): Promise<OnboardingQuestion[]> => {
    return client.get<OnboardingQuestion[]>('/v1/onboarding/questions')
  }

export const submitResponse = async (
    questionId: string, 
    responseValue: string | string[], 
    skipped = false
  ): Promise<void> => {
    const request: SubmitResponsesRequest = {
      responses: [{
        question_id: questionId,
        response_value: skipped ? undefined : responseValue,
        skipped
      }]
    }

    await client.post<{ message: string }>('/v1/onboarding/responses', request)
  }

export const skipQuestion = async (questionId: string): Promise<void> => {
    await client.post<{ message: string }>(`/onboarding/skip/${questionId}`, {})
  }

export const getStatus = async (): Promise<OnboardingStatus> => {
    return client.get<OnboardingStatus>('/v1/onboarding/status')
  }

export const submitMultipleResponses = async (responses: SubmitResponseRequest[]): Promise<void> => {
    const request: SubmitResponsesRequest = { responses }
    await client.post<{ message: string }>('/v1/onboarding/responses', request)
  }

export const isOnboardingComplete = async (): Promise<boolean> => {
    try {
      const status = await getStatus()
      return status.onboarding_completed
    } catch (error) {
      console.error('Error checking onboarding status:', error)
      return false
    }
  }

export const getNextQuestion = async (): Promise<OnboardingQuestion | null> => {
    const questions = await getQuestions()
    
    // Find first question that hasn't been answered or skipped
    const nextQuestion = questions.find(q => !q.user_answer && !q.is_skipped)
    
    return nextQuestion || null
  }

export const getCompletionPercentage = async (): Promise<number> => {
    const status = await getStatus()
    
    if (status.total_questions === 0) return 0
    
    const answeredOrSkipped = status.completed_questions + status.skipped_questions
    return Math.round((answeredOrSkipped / status.total_questions) * 100)
  }

