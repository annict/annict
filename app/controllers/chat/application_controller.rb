# frozen_string_literal: true

module Chat
  class ApplicationController < ActionController::Base
    include ControllerCommon
    include Localable
  end
end
