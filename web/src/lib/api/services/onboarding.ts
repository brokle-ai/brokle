import { BrokleAPIClient } from '../core/client'
import type {
  OnboardingQuestion,
  OnboardingStatus,
  SubmitResponseRequest,
  SubmitResponsesRequest
} from '@/types/onboarding'

export class OnboardingAPIClient extends BrokleAPIClient {
  constructor() {
    super('/auth') // All onboarding endpoints will be prefixed with /auth
  }

  /**
   * Fetch all active onboarding questions with user's current responses
   */
  async getQuestions(): Promise<OnboardingQuestion[]> {
    const response = await this.get<OnboardingQuestion[]>(
      '/v1/onboarding/questions'
    )
    return response
  }

  /**
   * Submit a single response for a question (individual submission)
   */
  async submitResponse(questionId: string, responseValue: string | string[], skipped = false): Promise<void> {
    const request: SubmitResponsesRequest = {
      responses: [{
        question_id: questionId,
        response_value: skipped ? undefined : responseValue,
        skipped
      }]
    }

    await this.post<{ message: string }>(
      '/v1/onboarding/responses',
      request
    )
  }

  /**
   * Skip an individual question
   */
  async skipQuestion(questionId: string): Promise<void> {
    await this.post<{ message: string }>(
      `/v1/onboarding/skip/${questionId}`,
      {}
    )
  }

  /**
   * Get current onboarding progress and status
   */
  async getStatus(): Promise<OnboardingStatus> {
    const response = await this.get<OnboardingStatus>(
      '/v1/onboarding/status'
    )
    return response
  }

  /**
   * Submit multiple responses at once (batch submission)
   * Keeping this for potential future use, but primary approach is individual submission
   */
  async submitMultipleResponses(responses: SubmitResponseRequest[]): Promise<void> {
    const request: SubmitResponsesRequest = { responses }
    
    await this.post<{ message: string }>(
      '/v1/onboarding/responses',
      request
    )
  }

  /**
   * Helper method to determine if onboarding is complete
   */
  async isOnboardingComplete(): Promise<boolean> {
    try {
      const status = await this.getStatus()
      return status.onboarding_completed
    } catch (error) {
      console.error('Error checking onboarding status:', error)
      return false
    }
  }

  /**
   * Get the next unanswered question based on current progress
   */
  async getNextQuestion(): Promise<OnboardingQuestion | null> {
    const questions = await this.getQuestions()
    
    // Find first question that hasn't been answered or skipped
    const nextQuestion = questions.find(q => !q.user_answer && !q.is_skipped)
    
    return nextQuestion || null
  }

  /**
   * Get completion percentage for progress tracking
   */
  async getCompletionPercentage(): Promise<number> {
    const status = await this.getStatus()
    
    if (status.total_questions === 0) return 0
    
    const answeredOrSkipped = status.completed_questions + status.skipped_questions
    return Math.round((answeredOrSkipped / status.total_questions) * 100)
  }
}