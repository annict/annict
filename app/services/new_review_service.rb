# frozen_string_literal: true

class NewReviewService
  attr_writer :app, :via, :ga_client, :keen_client, :page_category
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
      create_keen_event
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

  def create_keen_event
    return if @keen_client.blank?

    @keen_client.publish(
      "create_reviews",
      user: @user,
      page_category: @page_category,
      work_id: @review.work_id,
      via: @via,
      oauth_application_id: @app&.id
    )
  end
end
