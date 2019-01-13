# frozen_string_literal: true

class MenusController < ApplicationController
  def show
    return redirect_to root_path if device_pc?
  end
end
