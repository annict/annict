# frozen_string_literal: true

class NewRecordService
  attr_writer :app, :via, :ga_client, :keen_client, :page_category
  attr_reader :record

  def initialize(user, record)
    @user = user
    @record = record
  end

  def save!
    @record.user = @user
    @record.work = @record.episode.work

    ActiveRecord::Base.transaction do
      @record.save!
      @record.update_share_checkin_status
      @record.share_to_sns
      save_activity
      update_record_comments_count
      finish_tips
      update_latest_status
      create_ga_event
      create_keen_event
    end

    true
  end

  private

  def save_activity
    return if @record.multiple_record.present?
    CreateRecordActivityJob.perform_later(@user.id, @record.id)
  end

  def finish_tips
    FinishUserTipsJob.perform_later(@user, "checkin") if @user.records.initial?(@record)
  end

  def update_latest_status
    latest_status = @user.latest_statuses.find_by(work: @record.work)
    latest_status.append_episode(@record.episode) if latest_status.present?
  end

  def create_ga_event
    return if @ga_client.blank?
    data_source = @app.present? ? :api : :web
    @ga_client.events.create(:records, :create, ds: data_source)
  end

  def create_keen_event
    return if @keen_client.blank?

    @keen_client.publish(
      "create_records",
      user: @user,
      page_category: @page_category,
      work_id: @record.work_id,
      episode_id: @record.episode_id,
      has_comment: @record.comment.present?,
      shared_twitter: @record.shared_twitter?,
      shared_facebook: @record.shared_facebook?,
      is_first_record: @user.records.initial?(@record),
      via: @via,
      oauth_application_id: @app&.id
    )
  end

  def update_record_comments_count
    @record.episode.increment!(:record_comments_count) if @record.comment.present?
  end
end
