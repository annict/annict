# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for types exported from the `standard-sorbet` gem.
# Please instead update this file by running `bin/tapioca gem standard-sorbet`.


# source://standard-sorbet//lib/standard/sorbet/version.rb#1
module Standard; end

# source://standard-sorbet//lib/standard/sorbet/version.rb#2
module Standard::Sorbet; end

# source://standard-sorbet//lib/standard/sorbet/plugin.rb#2
class Standard::Sorbet::Plugin < ::LintRoller::Plugin
  # @return [Plugin] a new instance of Plugin
  #
  # source://standard-sorbet//lib/standard/sorbet/plugin.rb#3
  def initialize(config); end

  # source://standard-sorbet//lib/standard/sorbet/plugin.rb#8
  def about; end

  # source://standard-sorbet//lib/standard/sorbet/plugin.rb#21
  def rules(context); end

  # @return [Boolean]
  #
  # source://standard-sorbet//lib/standard/sorbet/plugin.rb#17
  def supported?(context); end

  private

  # This is not fantastic.
  #
  # When you `require "rubocop-sorbet"`, it will not only load the cops,
  # but it will also monkey-patch RuboCop's default_configuration, which is
  # something that can't be undone for the lifetime of the process.
  #
  # See: https://github.com/Shopify/rubocop-sorbet/blob/main/lib/rubocop/sorbet/inject.rb
  #
  # As an alternative, standard-sorbet loads the cops directly, and then
  # simply tells the RuboCop config loader that it's been loaded. This is
  # taking advantage of a private API of an `attr_reader` that probably wasn't
  # meant to be mutated externally, but it's better than the `Inject` monkey
  # patching that rubocop-sorbet does (and many other RuboCop plugins do)
  #
  # source://standard-sorbet//lib/standard/sorbet/plugin.rb#51
  def trick_rubocop_into_thinking_we_required_rubocop_sorbet!; end
end

# source://standard-sorbet//lib/standard/sorbet/version.rb#3
Standard::Sorbet::VERSION = T.let(T.unsafe(nil), String)
