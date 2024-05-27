# typed: false
# frozen_string_literal: true

# Remove Rails field_with_errors wrapper to style with Bootstrap
# https://coderwall.com/p/s-zwrg/remove-rails-field_with_errors-wrapper
ActionView::Base.field_error_proc = proc do |html_tag|
  html_tag.html_safe
end
