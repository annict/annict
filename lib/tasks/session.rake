# typed: false
# frozen_string_literal: true

namespace :session do
  task sweep: :environment do
    Session.where("updated_at < '#{10.days.ago.to_s(:db)}'").delete_all
  end
end
