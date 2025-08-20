<script setup lang="ts">
defineProps<{
  open: boolean
  disableButtons?: boolean
}>()
defineEmits(['confirm', 'cancel'])
</script>

<template>
  <AlertDialog :open="open">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle><slot name="title" /></AlertDialogTitle>
        <AlertDialogDescription>
          <slot name="description" />
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <slot name="footer" />
        <AlertDialogCancel v-if="!disableButtons" @click="$emit('cancel')">
          Cancel
        </AlertDialogCancel>
        <AlertDialogAction v-if="!disableButtons" type="submit" @click="$emit('confirm')">
          OK
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
