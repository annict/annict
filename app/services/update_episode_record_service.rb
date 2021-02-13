# frozen_string_literal: true

class UpdateEpisodeRecordService < ApplicationService
  class ServiceResult < Result
    attr_accessor :record
  end

  def initialize(
    user:,
    record:,
    rating: nil,
    comment: "",
    share_to_twitter: false,
    oauth_application: nil
  ) # rubocop:disable Metrics/ParameterLists
    super()
    @user = user
    @record = record
    @episode_record = @record.episode_record
    @rating = rating
    @comment = comment
    @share_to_twitter = share_to_twitter
    @oauth_application = oauth_application
  end

  def call
    @episode_record.rating_state = @rating&.downcase
    @episode_record.modify_body = @episode_record.body != @comment
    @episode_record.body = @comment
    @episode_record.oauth_application = @oauth_application
    @episode_record.detect_locale!(:body)

    ActiveRecord::Base.transaction do
      @episode_record.save!
      @user.touch(:record_cache_expired_at)

      if @share_to_twitter
        @user.share_episode_record_to_twitter(@episode_record)
      end
    end

    if @result.success?
      @result.record = @record
    end

    @result
  end

  private

  def result_class
    ServiceResult
  end
end
