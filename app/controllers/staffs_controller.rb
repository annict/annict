# frozen_string_literal: true

class StaffsController < ApplicationController
  before_action :load_work, only: %i(index)

  def index
    @staffs = @work.staffs.published.order(:sort_number)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end
end
