# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for types exported from the `puma_worker_killer` gem.
# Please instead update this file by running `bin/tapioca gem puma_worker_killer`.


# source://puma_worker_killer//lib/puma_worker_killer.rb#5
module PumaWorkerKiller
  extend ::PumaWorkerKiller

  # @yield [_self]
  # @yieldparam _self [PumaWorkerKiller] the object that the method was called on
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#22
  def config; end

  # source://puma_worker_killer//lib/puma_worker_killer.rb#35
  def enable_rolling_restart(frequency = T.unsafe(nil), splay_seconds = T.unsafe(nil)); end

  # Returns the value of attribute frequency.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def frequency; end

  # Sets the attribute frequency
  #
  # @param value the value to set the attribute frequency to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def frequency=(_arg0); end

  # Returns the value of attribute on_calculation.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def on_calculation; end

  # Sets the attribute on_calculation
  #
  # @param value the value to set the attribute on_calculation to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def on_calculation=(_arg0); end

  # Returns the value of attribute percent_usage.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def percent_usage; end

  # Sets the attribute percent_usage
  #
  # @param value the value to set the attribute percent_usage to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def percent_usage=(_arg0); end

  # Returns the value of attribute pre_term.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def pre_term; end

  # Sets the attribute pre_term
  #
  # @param value the value to set the attribute pre_term to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def pre_term=(_arg0); end

  # Returns the value of attribute ram.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def ram; end

  # Sets the attribute ram
  #
  # @param value the value to set the attribute ram to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def ram=(_arg0); end

  # source://puma_worker_killer//lib/puma_worker_killer.rb#26
  def reaper(ram = T.unsafe(nil), percent_usage = T.unsafe(nil), reaper_status_logs = T.unsafe(nil), pre_term = T.unsafe(nil), on_calculation = T.unsafe(nil)); end

  # Returns the value of attribute reaper_status_logs.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def reaper_status_logs; end

  # Sets the attribute reaper_status_logs
  #
  # @param value the value to set the attribute reaper_status_logs to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def reaper_status_logs=(_arg0); end

  # Returns the value of attribute rolling_pre_term.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def rolling_pre_term; end

  # Sets the attribute rolling_pre_term
  #
  # @param value the value to set the attribute rolling_pre_term to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def rolling_pre_term=(_arg0); end

  # Returns the value of attribute rolling_restart_frequency.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def rolling_restart_frequency; end

  # Sets the attribute rolling_restart_frequency
  #
  # @param value the value to set the attribute rolling_restart_frequency to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def rolling_restart_frequency=(_arg0); end

  # Returns the value of attribute rolling_restart_splay_seconds.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def rolling_restart_splay_seconds; end

  # Sets the attribute rolling_restart_splay_seconds
  #
  # @param value the value to set the attribute rolling_restart_splay_seconds to.
  #
  # source://puma_worker_killer//lib/puma_worker_killer.rb#8
  def rolling_restart_splay_seconds=(_arg0); end

  # source://puma_worker_killer//lib/puma_worker_killer.rb#30
  def start(frequency = T.unsafe(nil), reaper = T.unsafe(nil)); end
end

# source://puma_worker_killer//lib/puma_worker_killer/auto_reap.rb#4
class PumaWorkerKiller::AutoReap
  # @return [AutoReap] a new instance of AutoReap
  #
  # source://puma_worker_killer//lib/puma_worker_killer/auto_reap.rb#5
  def initialize(timeout, reaper = T.unsafe(nil)); end

  # source://puma_worker_killer//lib/puma_worker_killer/auto_reap.rb#11
  def start; end
end

# source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#4
class PumaWorkerKiller::PumaMemory
  # @return [PumaMemory] a new instance of PumaMemory
  #
  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#5
  def initialize(master = T.unsafe(nil)); end

  # Will refresh @workers
  #
  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#53
  def get_total(workers = T.unsafe(nil)); end

  # Will refresh @workers
  #
  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#53
  def get_total_memory(workers = T.unsafe(nil)); end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#42
  def largest_worker; end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#47
  def largest_worker_memory; end

  # Returns the value of attribute master.
  #
  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#10
  def master; end

  # @return [Boolean]
  #
  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#28
  def running?; end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#12
  def size; end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#32
  def smallest_worker; end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#37
  def smallest_worker_memory; end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#20
  def term_largest_worker; end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#16
  def term_worker(worker); end

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#60
  def workers; end

  # @return [Boolean]
  #
  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#24
  def workers_stopped?; end

  private

  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#66
  def get_master; end

  # Returns sorted hash, keys are worker objects, values are memory used per worker
  # sorted by memory ascending (smallest first, largest last)
  #
  # source://puma_worker_killer//lib/puma_worker_killer/puma_memory.rb#72
  def set_workers; end
end

# source://puma_worker_killer//lib/puma_worker_killer/reaper.rb#4
class PumaWorkerKiller::Reaper
  # @return [Reaper] a new instance of Reaper
  #
  # source://puma_worker_killer//lib/puma_worker_killer/reaper.rb#5
  def initialize(max_ram, master = T.unsafe(nil), reaper_status_logs = T.unsafe(nil), pre_term = T.unsafe(nil), on_calculation = T.unsafe(nil)); end

  # used for tes
  #
  # source://puma_worker_killer//lib/puma_worker_killer/reaper.rb#14
  def get_total_memory; end

  # source://puma_worker_killer//lib/puma_worker_killer/reaper.rb#18
  def reap; end
end

# source://puma_worker_killer//lib/puma_worker_killer/rolling_restart.rb#4
class PumaWorkerKiller::RollingRestart
  # @return [RollingRestart] a new instance of RollingRestart
  #
  # source://puma_worker_killer//lib/puma_worker_killer/rolling_restart.rb#5
  def initialize(master = T.unsafe(nil), rolling_pre_term = T.unsafe(nil)); end

  # used for tes
  #
  # source://puma_worker_killer//lib/puma_worker_killer/rolling_restart.rb#11
  def get_total_memory; end

  # source://puma_worker_killer//lib/puma_worker_killer/rolling_restart.rb#15
  def reap(seconds_between_worker_kill = T.unsafe(nil)); end
end

# source://puma_worker_killer//lib/puma_worker_killer/version.rb#4
PumaWorkerKiller::VERSION = T.let(T.unsafe(nil), String)
