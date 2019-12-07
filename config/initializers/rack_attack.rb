# frozen_string_literal: true

BLOCK_USER_AGENTS = %w(
  BLEXBot
).freeze

BLOCK_IPS = %w(
  144.76.0.0/16
  216.244.66.0/24
  94.130.0.0/16
).freeze

def remote_ip(req)
  @remote_ip ||= (req.env["action_dispatch.remote_ip"] || req.ip).to_s
end

def include_ua?(req)
  BLOCK_USER_AGENTS.any? { |ua| req.user_agent&.include?(ua) }
end

def include_ip?(req)
  BLOCK_IPS.any? { |ip| IPAddr.new(ip).include?(remote_ip(req)) }
end

Rack::Attack.blocklist("block all access from specific user agents and ips") do |req|
  include_ua?(req) || include_ip?(req)
end

Rack::Attack.throttle("requests by ip", limit: 5, period: 2.seconds) do |req|
  remote_ip(req)
end

ActiveSupport::Notifications.subscribe(/rack_attack/) do |_name, _start, _finish, _request_id, payload|
  req = payload[:request]

  next unless %i(throttle blacklist).include? req.env["rack.attack.match_type"]

  Rails.logger.info("Rate limit hit (#{req.env['rack.attack.match_type']}): #{remote_ip(req)} #{req.request_method} #{req.fullpath}")
end
