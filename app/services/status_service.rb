# frozen_string_literal: true

class StatusService
  attr_writer :app, :keen_client, :ga_client

  def initialize(user, work)
    @user = user
    @work = work
  end

  def change!(kind)
    if Status.kind.values.include?(kind)
      status = @user.statuses.new(work: @work, kind: kind, oauth_application: @app)

      if status.save!
        UserWatchedWorksCountJob.perform_later(@user)
        @keen_client.app = @app
        @keen_client.statuses.create
        data_source = @app.present? ? :api : :web
        @ga_client.events.create(:statuses, :create, ds: data_source)
      end
    elsif kind == "no_select"
      latest_status = @user.latest_statuses.find_by(work: @work)
      if latest_status.present?
        latest_status.destroy!
        UserWatchedWorksCountJob.perform_later(@user)
      end
    end
  end
end
