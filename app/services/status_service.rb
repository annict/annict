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
        data_source = @app.present? ? :api : :web
        @ga_client.events.create(:statuses, :create, ds: data_source)
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

  def create_keen_event
    return if @keen_client.blank?

    @keen_client.publish(
      "create_statuses",
      user: @user,
      page_category: @page_category,
      work_id: @work.id,
      kind: @kind,
      is_first_status: @user.statuses.initial?(@status),
      via: @via,
      oauth_application_id: @app&.id
    )
  end
end
