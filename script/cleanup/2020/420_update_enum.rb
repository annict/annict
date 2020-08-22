# frozen_string_literal: true

from = Time.zone.parse(ARGV[0])

KIND_MAPPING = {
  wanna_watch: :want_to_watch,
  watching: :watching,
  watched: :completed,
  on_hold: :on_hold,
  stop_watching: :dropped
}.freeze

Status.after(from, field: :updated_at).find_each do |status|
  puts "status: #{status.id}"
  status.update_column(:new_kind, KIND_MAPPING[status.kind.to_sym])
end
