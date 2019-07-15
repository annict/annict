# frozen_string_literal: true

module Chat
  class ApplicationController < ActionController::Base
    include ControllerCommon
    include Localable

    helper_method :locale_ja?, :locale_en?
  end
end
