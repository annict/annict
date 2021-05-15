# frozen_string_literal: true

module V6::Fragment
  class ApplicationController < ActionController::Base
    include V6::PageCategorizable

    layout false
  end
end
