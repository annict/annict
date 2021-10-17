# frozen_string_literal: true

module Creators
  class WorkRecordCreator
    attr_accessor :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @work = @form.work
    end

    def call
      work_record = @work.build_work_record(
        user: @user,
        rating_overall: @form.rating_overall,
        rating_animation: @form.rating_animation,
        rating_music: @form.rating_music,
        rating_story: @form.rating_story,
        rating_character: @form.rating_character,
        comment: @form.comment,
        share_to_twitter: @form.share_to_twitter,
        watched_at: @form.watched_at
      )

      if @form.deprecated_title.present?
        work_record.body = "#{@form.deprecated_title}\n\n#{work_record.body}"
      end

      ActiveRecord::Base.transaction do
        work_record.save!

        if @form.create_activity?
          activity_group = @user.create_or_last_activity_group!(work_record)
          @user.activities.create!(itemable: work_record, activity_group: activity_group)
        end

        @user.update_share_record_setting(@form.share_to_twitter)
        @user.touch(:record_cache_expired_at)

        if @user.share_record_to_twitter?
          @user.share_work_record_to_twitter(work_record)
        end
      end

      self.record = work_record.record

      self
    end
  end
end
