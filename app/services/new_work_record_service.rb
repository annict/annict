# frozen_string_literal: true

class NewWorkRecordService
  attr_writer :app, :via, :ga_client, :timber, :page_category
  attr_reader :work_record

  def initialize(user, work_record, setting)
    @user = user
    @work_record = work_record
    @work = @work_record.work
    @setting = setting
  end

  def save!
    ActiveRecord::Base.transaction do
      @work_record.detect_locale!(:body)
      @work_record.record = @user.records.create!(work: @work)

      @work_record.save!
      @setting.save!
    end

    @work_record.share_to_sns
    save_activity
    create_ga_event
    create_timber_log

    true
  end

  private

  def save_activity
    CreateWorkRecordActivityJob.perform_later(@user.id, @work_record.id)
  end

  def create_ga_event
    return if @ga_client.blank?
    @ga_client.events.create(:records, :create, el: "Work", ds: @via)
  end

  def create_timber_log
    return if @timber.blank?
    @timber.log(:info, :WORK_RECORD_CREATE, via: @via)
  end
end
