# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for types exported from the `commonmarker` gem.
# Please instead update this file by running `bin/tapioca gem commonmarker`.


# source://commonmarker//lib/commonmarker/constants.rb#3
module Commonmarker
  private

  def commonmark_parse(*_arg0); end
  def commonmark_to_html(*_arg0); end

  class << self
    def commonmark_parse(*_arg0); end
    def commonmark_to_html(*_arg0); end

    # Public: Parses a CommonMark string into an HTML string.
    #
    # text - A {String} of text
    # options - A {Hash} of render, parse, and extension options to transform the text.
    #
    # Returns the `parser` node.
    #
    # @raise [TypeError]
    #
    # source://commonmarker//lib/commonmarker.rb#19
    def parse(text, options: T.unsafe(nil)); end

    # Public: Parses a CommonMark string into an HTML string.
    #
    # text - A {String} of text
    # options - A {Hash} of render, parse, and extension options to transform the text.
    # plugins - A {Hash} of additional plugins.
    #
    # Returns a {String} of converted HTML.
    #
    # @raise [TypeError]
    #
    # source://commonmarker//lib/commonmarker.rb#36
    def to_html(text, options: T.unsafe(nil), plugins: T.unsafe(nil)); end
  end
end

# source://commonmarker//lib/commonmarker/config.rb#4
module Commonmarker::Config
  extend ::Commonmarker::Constants
  extend ::Commonmarker::Utils

  class << self
    # source://commonmarker//lib/commonmarker/config.rb#81
    def process_extension_options(options); end

    # source://commonmarker//lib/commonmarker/config.rb#65
    def process_options(options); end

    # source://commonmarker//lib/commonmarker/config.rb#81
    def process_parse_options(options); end

    # source://commonmarker//lib/commonmarker/config.rb#73
    def process_plugins(plugins); end

    # source://commonmarker//lib/commonmarker/config.rb#81
    def process_render_options(options); end

    # source://commonmarker//lib/commonmarker/config.rb#98
    def process_syntax_highlighter_plugin(options); end
  end
end

# For details, see
# https://github.com/kivikakk/comrak/blob/162ef9354deb2c9b4a4e05be495aa372ba5bb696/src/main.rs#L201
#
# source://commonmarker//lib/commonmarker/config.rb#7
Commonmarker::Config::OPTIONS = T.let(T.unsafe(nil), Hash)

# source://commonmarker//lib/commonmarker/config.rb#55
Commonmarker::Config::PLUGINS = T.let(T.unsafe(nil), Hash)

# source://commonmarker//lib/commonmarker/constants.rb#4
module Commonmarker::Constants; end

# source://commonmarker//lib/commonmarker/constants.rb#5
Commonmarker::Constants::BOOLS = T.let(T.unsafe(nil), Array)

# source://commonmarker//lib/commonmarker/node/ast.rb#4
class Commonmarker::Node
  include ::Prelude::Enumerator
  include ::Enumerable
  include ::Commonmarker::Node::Inspect

  def append_child(_arg0); end
  def delete; end

  # Public: Iterate over the children (if any) of the current pointer.
  #
  # source://commonmarker//lib/commonmarker/node.rb#24
  def each; end

  def fence_info; end
  def fence_info=(_arg0); end
  def first_child; end
  def header_level; end
  def header_level=(_arg0); end
  def insert_after(_arg0); end
  def insert_before(_arg0); end
  def last_child; end
  def list_start; end
  def list_start=(_arg0); end
  def list_tight; end
  def list_tight=(_arg0); end
  def list_type; end
  def list_type=(_arg0); end
  def next_sibling; end
  def node_to_commonmark(*_arg0); end
  def node_to_html(*_arg0); end
  def parent; end
  def prepend_child(_arg0); end
  def previous_sibling; end
  def replace(_arg0); end
  def source_position; end
  def string_content; end
  def string_content=(_arg0); end
  def title; end
  def title=(_arg0); end

  # Public: Convert the node to a CommonMark string.
  #
  # options - A {Symbol} or {Array of Symbol}s indicating the render options
  # plugins - A {Hash} of additional plugins.
  #
  # Returns a {String}.
  #
  # @raise [TypeError]
  #
  # source://commonmarker//lib/commonmarker/node.rb#56
  def to_commonmark(options: T.unsafe(nil), plugins: T.unsafe(nil)); end

  # Public: Converts a node to an HTML string.
  #
  # options - A {Hash} of render, parse, and extension options to transform the text.
  # plugins - A {Hash} of additional plugins.
  #
  # Returns a {String} of HTML.
  #
  # @raise [TypeError]
  #
  # source://commonmarker//lib/commonmarker/node.rb#41
  def to_html(options: T.unsafe(nil), plugins: T.unsafe(nil)); end

  def type; end
  def url; end
  def url=(_arg0); end

  # Public: An iterator that "walks the tree," descending into children recursively.
  #
  # blk - A {Proc} representing the action to take for each child
  #
  # @yield [_self]
  # @yieldparam _self [Commonmarker::Node] the object that the method was called on
  #
  # source://commonmarker//lib/commonmarker/node.rb#14
  def walk(&block); end

  class << self
    def new(*_arg0); end
  end
end

# source://commonmarker//lib/commonmarker/node/ast.rb#5
class Commonmarker::Node::Ast; end

# source://commonmarker//lib/commonmarker/node/inspect.rb#7
module Commonmarker::Node::Inspect
  # source://commonmarker//lib/commonmarker/node/inspect.rb#10
  def inspect; end

  # @param printer [PrettyPrint] pp
  #
  # source://commonmarker//lib/commonmarker/node/inspect.rb#15
  def pretty_print(printer); end
end

# source://commonmarker//lib/commonmarker/node/inspect.rb#8
Commonmarker::Node::Inspect::PP_INDENT_SIZE = T.let(T.unsafe(nil), Integer)

# source://commonmarker//lib/commonmarker/renderer.rb#7
class Commonmarker::Renderer; end

# source://commonmarker//lib/commonmarker/utils.rb#6
module Commonmarker::Utils
  include ::Commonmarker::Constants

  # source://commonmarker//lib/commonmarker/utils.rb#9
  def fetch_kv(options, key, value, type); end
end

# source://commonmarker//lib/commonmarker/version.rb#4
Commonmarker::VERSION = T.let(T.unsafe(nil), String)
