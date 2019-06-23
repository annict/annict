import 'bootstrap'
import Turbolinks from 'turbolinks'
import { Application } from 'stimulus'
import { definitionsFromContext } from 'stimulus/webpack-helpers'

const application = Application.start()
const context = require.context('./v3/controllers', true, /\.js$/)
application.load(definitionsFromContext(context))

Turbolinks.start()
