# frozen_string_literal: true

module ControllerCommon
  extend ActiveSupport::Concern

  included do
    helper_method :render_jb

    def render_jb(path, assigns)
      ApplicationController.render("#{path}.jb", assigns: assigns)
    end
  end
end
