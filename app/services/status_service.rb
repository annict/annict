# frozen_string_literal: true

class StatusService
  attr_writer :app, :via, :ga_client, :keen_client, :page_category

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
        create_keen_event
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

  def create_keen_event
    return if @keen_client.blank?
    @keen_client.publish(
      :status_create,
      via: @via,
      kind: @kind,
      is_first_status: @user.statuses.initial?(@status),
      oauth_application_id: @app&.id
    )
  end
end
