class SettingsController < ApplicationController
  before_filter :authenticate_user!
end
