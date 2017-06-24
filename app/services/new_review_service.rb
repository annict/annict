# frozen_string_literal: true

class NewReviewService
  attr_writer :app, :ga_client
  attr_reader :review

  def initialize(user, review, setting)
    @user = user
    @review = review
    @setting = setting
  end

  def save!
    ActiveRecord::Base.transaction do
      @review.save!
      @setting.save!
      @review.share_to_sns
      save_activity
      create_ga_event
    end

    true
  end

  private

  def save_activity
    CreateReviewActivityJob.perform_later(@user.id, @review.id)
  end

  def create_ga_event
    return if @ga_client.blank?
    data_source = @app.present? ? :api : :web
    @ga_client.events.create(:reviews, :create, ds: data_source)
  end
end
