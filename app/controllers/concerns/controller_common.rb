# frozen_string_literal: true

module ControllerCommon
  extend ActiveSupport::Concern

  included do
    def render_jb(path, assigns)
      ApplicationController.render("#{path}.jb", assigns: assigns)
    end
  end
end
