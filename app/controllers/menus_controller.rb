# frozen_string_literal: true

class MenusController < ApplicationController
  def show
    return redirect_to root_path unless browser.device.mobile?
  end
end
