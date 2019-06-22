# This file is autogenerated. Do not edit it by hand. Regenerate it with:
#   srb rbi gems

# typed: true
#
# If you would like to make changes to this file, great! Please create the gem's shim here:
#
#   https://github.com/sorbet/sorbet-typed/new/master?filename=lib/dotenv/all/dotenv.rbi
#
# dotenv-5c812db9cfce
module Dotenv
  def ignoring_nonexistent_files; end
  def instrument(name, payload = nil, &block); end
  def load!(*filenames); end
  def load(*filenames); end
  def overload!(*filenames); end
  def overload(*filenames); end
  def parse(*filenames); end
  def require_keys(*keys); end
  def self.ignoring_nonexistent_files; end
  def self.instrument(name, payload = nil, &block); end
  def self.instrumenter; end
  def self.instrumenter=(arg0); end
  def self.load!(*filenames); end
  def self.load(*filenames); end
  def self.overload!(*filenames); end
  def self.overload(*filenames); end
  def self.parse(*filenames); end
  def self.require_keys(*keys); end
  def self.with(*filenames); end
  def with(*filenames); end
end
module Dotenv::Substitutions
end
module Dotenv::Substitutions::Variable
  def self.call(value, env, is_load); end
  def self.substitute(match, variable, env); end
end
module Dotenv::Substitutions::Command
  def self.call(value, _env, _is_load); end
end
class Dotenv::FormatError < SyntaxError
end
class Dotenv::Parser
  def call; end
  def expand_newlines(value); end
  def initialize(string, is_load = nil); end
  def parse_line(line); end
  def parse_value(value); end
  def self.call(string, is_load = nil); end
  def self.substitutions; end
  def unescape_characters(value); end
  def variable_not_set?(line); end
end
class Dotenv::Environment < Hash
  def apply!; end
  def apply; end
  def filename; end
  def initialize(filename, is_load = nil); end
  def load(is_load = nil); end
  def read; end
end
class Dotenv::Error < StandardError
end
class Dotenv::MissingKeys < Dotenv::Error
  def initialize(keys); end
end
class Dotenv::Railtie < Rails::Railtie
  def dotenv_files; end
  def load; end
  def root; end
  def self.load; end
end
