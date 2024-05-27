# typed: false
# frozen_string_literal: true

namespace :sorbet do
  task update: :environment do
    system "bin/tapioca gem"
    system "bin/tapioca dsl"
    system "bin/tapioca todo"
    system "bin/tapioca annotations"
  end
end
