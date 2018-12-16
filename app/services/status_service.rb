# frozen_string_literal: true

class StatusService
  attr_writer :app, :via, :ga_client, :logentries, :page_category

  def initialize(user, work)
    @user = user
    @work = work
  end

  def change!(kind)
    @kind = kind

    if Status.kind.values.include?(@kind)
      @status = @user.statuses.new(work: @work, kind: @kind, oauth_application: @app)

      if @status.save!
        UserWatchedWorksCountJob.perform_later(@user)
        @status.share_to_sns
        create_ga_event
        create_logentries_log
      end
    elsif @kind == "no_select"
      latest_status = @user.latest_statuses.find_by(work: @work)
      if latest_status.present?
        latest_status.destroy!
        UserWatchedWorksCountJob.perform_later(@user)
      end
    end
  end

  private

  def create_ga_event
    return if @ga_client.blank?
    @ga_client.events.create(:statuses, :create, ds: @via)
  end

  def create_logentries_log
    return if @logentries.blank?
    @logentries.log(:info, :STATUS_CREATE, via: @via)
  end
end
