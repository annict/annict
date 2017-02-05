# frozen_string_literal: true

class NewRecordService
  attr_writer :app
  attr_reader :record

  def initialize(user, record, keen_client)
    @user = user
    @record = record
    @keen_client = keen_client
  end

  def save
    @record.user = @user
    @record.work = @record.episode.work

    return false unless @record.valid?

    ActiveRecord::Base.transaction do
      @record.save(validate: false)
      @record.update_share_checkin_status
      @record.share_to_sns
      save_activity
      finish_tips
      update_latest_status
      create_keen_event
    end

    true
  end

  private

  def save_activity
    if @record.multiple_record.blank?
      Activity.delay.create(
        user: @user,
        recipient: @record.episode,
        trackable: @record,
        action: "create_record")
    end
  end

  def finish_tips
    UserTipsService.new(@user).delay.finish!(:checkin) if @user.records.initial?(@record)
  end

  def update_latest_status
    latest_status = @user.latest_statuses.find_by(work: @record.work)
    latest_status.append_episode(@record.episode) if latest_status.present?
  end

  def create_keen_event
    @keen_client.app = @app
    @keen_client.records.create
  end
end
