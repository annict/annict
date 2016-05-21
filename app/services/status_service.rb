# frozen_string_literal: true

class StatusService
  attr_writer :app

  def initialize(user, work, ga_client)
    @user = user
    @work = work
    @ga_client = ga_client
  end

  def change(kind)
    if Status.kind.values.include?(kind)
      status = @user.statuses.new(work: @work, kind: kind, oauth_application: @app)

      if status.save
        create_ga_event
        return true
      end
    elsif kind == "no_select"
      latest_status = @user.latest_statuses.find_by(work: @work)
      latest_status.destroy! if latest_status.present?
      return true
    end

    false
  end

  private

  def create_ga_event
    ds = @app.present? ? "app-#{@app.uid}" : "web"
    @ga_client.events.create("statuses", "create", ds: ds)
  end
end
