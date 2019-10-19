# frozen_string_literal: true

class CheckinsController < ApplicationController
  # Old record page
  def show
    record = Record.published.find(params[:id])
    redirect_to record_path(record.user.username, record), status: 301
  end
end
