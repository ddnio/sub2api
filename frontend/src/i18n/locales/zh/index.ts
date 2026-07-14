import landing from './landing'
import docs from './docs'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'

export default {
  ...landing,
  docs,
  ...common,
  ...dashboard,
  admin,
  ...misc,
}
