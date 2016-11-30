# frozen_string_literal: true

class StaffsController < ApplicationController
  before_action :load_work, only: %i(index)

  def index
    @staffs = @work.staffs.published.order(:sort_number)
  end
end
