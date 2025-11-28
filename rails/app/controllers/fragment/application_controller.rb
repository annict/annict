# typed: false
# frozen_string_literal: true

module Fragment
  class ApplicationController < ActionController::Base
    include PageCategorizable

    layout false
  end
end
