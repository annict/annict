module Api
  class WorksController < Api::ApplicationController
    def friends(user_id, work_id)
      user = User.find(user_id)
      work = Work.find(work_id)

      @users = user.friends_interested_in(work).
        includes(:profile).
        order("latest_statuses.id DESC")
    end
  end
end
