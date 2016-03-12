# frozen_string_literal: true

module Api
  class WorksController < Api::ApplicationController
    before_action :authenticate_user!

    def friends(work_id)
      work = Work.find(work_id)

      @users = current_user.friends_interested_in(work).
        includes(:profile).
        order("latest_statuses.id DESC")
    end
  end
end
